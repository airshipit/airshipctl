/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package replacement

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document/plugin/kyamlutils"
)

var (
	// substring substitutions are appended to paths as: ...%VARNAME%
	substringPatternRegex = regexp.MustCompile(`(.+)%(\S+)%$`)
)

const (
	secret     = "Secret"
	secretData = "data"
)

var _ kio.Filter = &plugin{}

type plugin struct {
	*airshipv1.ReplacementTransformer
}

// New creates new instance of the plugin
func New(obj map[string]interface{}) (kio.Filter, error) {
	cfg := &airshipv1.ReplacementTransformer{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, cfg)
	if err != nil {
		return nil, err
	}
	p := &plugin{ReplacementTransformer: cfg}
	for _, r := range p.Replacements {
		if r.Source == nil {
			return nil, ErrBadConfiguration{Msg: "`from` must be specified in one replacement"}
		}
		if r.Target == nil {
			return nil, ErrBadConfiguration{Msg: "`to` must be specified in one replacement"}
		}
		if r.Source.ObjRef != nil && r.Source.Value != "" {
			return nil, ErrBadConfiguration{Msg: "only one of fieldref and value is allowed in one replacement"}
		}
	}
	return p, nil
}

func (p *plugin) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, r := range p.Replacements {
		val, err := getValue(items, r.Source)
		if err != nil {
			return nil, err
		}

		if err := replace(items, r.Target, val); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func getValue(items []*yaml.RNode, source *types.ReplSource) (*yaml.RNode, error) {
	if source.Value != "" {
		return yaml.NewScalarRNode(source.Value), nil
	}
	sources, err := kyamlutils.DocumentSelector{}.
		ByAPIVersion(source.ObjRef.APIVersion).
		ByGVK(source.ObjRef.Group, source.ObjRef.Version, source.ObjRef.Kind).
		ByName(source.ObjRef.Name).
		ByNamespace(source.ObjRef.Namespace).
		Filter(items)
	if err != nil {
		return nil, err
	}

	if len(sources) > 1 {
		return nil, ErrMultipleResources{ObjRef: source.ObjRef}
	}
	if len(sources) == 0 {
		return nil, ErrSourceNotFound{ObjRef: source.ObjRef}
	}

	path := fmt.Sprintf("{.%s.%s}", yaml.MetadataField, yaml.NameField)
	if source.FieldRef != "" {
		path = source.FieldRef
	}
	sourceNode, err := sources[0].Pipe(kyamlutils.JSONPathFilter{Path: path})
	if err != nil {
		return nil, err
	}

	// Decoding value if source is `kind: Secret` and
	// has fieldRef `data`
	if source.ObjRef.Gvk.Kind == secret && strings.Split(source.FieldRef, ".")[0] == secretData {
		return decodeValue(sourceNode)
	}
	return sourceNode, nil
}

func mutateField(rnSource *yaml.RNode) func([]*yaml.RNode) error {
	return func(rns []*yaml.RNode) error {
		for _, rn := range rns {
			rn.SetYNode(rnSource.YNode())
		}
		return nil
	}
}

func replace(items []*yaml.RNode, target *types.ReplTarget, value *yaml.RNode) error {
	targets, err := kyamlutils.DocumentSelector{}.
		ByGVK(target.ObjRef.Group, target.ObjRef.Version, target.ObjRef.Kind).
		ByName(target.ObjRef.Name).
		ByNamespace(target.ObjRef.Namespace).
		Filter(items)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return ErrTargetNotFound{ObjRef: target.ObjRef}
	}
	for _, tgt := range targets {
		for _, fieldRef := range target.FieldRefs {
			val := value
			// Encoding value before replacement if target is `kind: Secret`
			// and has fieldRef `data`
			if target.ObjRef.Gvk.Kind == secret && strings.Split(fieldRef, ".")[0] == secretData {
				val = encodeValue(val)
			}
			// fieldref can contain substring pattern for regexp - we need to get it
			groups := substringPatternRegex.FindStringSubmatch(fieldRef)
			// if there is no substring pattern
			if len(groups) != 3 {
				filter := kyamlutils.JSONPathFilter{Path: fieldRef, Mutator: mutateField(val), Create: true}
				if _, err := tgt.Pipe(filter); err != nil {
					return err
				}
				continue
			}

			if err := substituteSubstring(tgt, groups[1], groups[2], val); err != nil {
				return err
			}
		}
	}
	return nil
}

func substituteSubstring(tgt *yaml.RNode, fieldRef, substringPattern string, value *yaml.RNode) error {
	if err := yaml.ErrorIfInvalid(value, yaml.ScalarNode); err != nil {
		return err
	}
	curVal, err := tgt.Pipe(kyamlutils.JSONPathFilter{Path: fieldRef})
	if yaml.IsMissingOrError(curVal, err) {
		return err
	}
	switch curVal.YNode().Kind {
	case yaml.ScalarNode:
		p := regexp.MustCompile(substringPattern)
		curVal.YNode().Value = p.ReplaceAllString(yaml.GetValue(curVal), yaml.GetValue(value))

	case yaml.SequenceNode:
		items, err := curVal.Elements()
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := yaml.ErrorIfInvalid(item, yaml.ScalarNode); err != nil {
				return err
			}
			p := regexp.MustCompile(substringPattern)
			item.YNode().Value = p.ReplaceAllString(yaml.GetValue(item), yaml.GetValue(value))
		}
	default:
		return ErrPatternSubstring{Msg: fmt.Sprintf("value identified by %s expected to be string", fieldRef)}
	}
	return nil
}

func decodeValue(source *yaml.RNode) (*yaml.RNode, error) {
	//decoding replacement source if source reference has data field
	decodeValue, err := decodeString(yaml.GetValue(source))
	if err != nil {
		return nil, ErrBase64Decoding{Msg: fmt.Sprintf("Error while decoding base64"+
			" encoded string: %s", yaml.GetValue(source))}
	}
	node := yaml.NewScalarRNode(decodeValue)
	return node, nil
}

func encodeValue(value *yaml.RNode) *yaml.RNode {
	encodeValue := encodeString(yaml.GetValue(value))
	node := yaml.NewScalarRNode(encodeValue)
	return node
}

func decodeString(value interface{}) (string, error) {
	decodedValue, err := base64.StdEncoding.DecodeString(value.(string))
	if err != nil {
		return "", err
	}
	return string(decodedValue), nil
}

func encodeString(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

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

package kyamlutils

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var _ kio.Filter = DocumentSelector{}

// DocumentSelector RNode objects
type DocumentSelector struct {
	filters []kio.Filter
}

// Filters return list of defined filters for the selector
func (f DocumentSelector) Filters() []kio.Filter {
	return f.filters
}

func (f DocumentSelector) byPath(path []string, val string) DocumentSelector {
	// Need to have exact match of the value since grep filter considers Value
	// as a regular expression
	f.filters = append(f.filters, filters.GrepFilter{Path: path, Value: "^" + val + "$"})
	return f
}

// ByKey adds filter by specific yaml manifest key and value
func (f DocumentSelector) ByKey(key, val string) DocumentSelector {
	return f.byPath([]string{key}, val)
}

// ByAPIVersion adds filter by 'apiVersion' field value
func (f DocumentSelector) ByAPIVersion(apiver string) DocumentSelector {
	if apiver != "" {
		return f.ByKey(yaml.APIVersionField, apiver)
	}
	return f
}

// ByGVK adds filters by 'apiVersion' and 'kind; field values
func (f DocumentSelector) ByGVK(group, version, kind string) DocumentSelector {
	apiver := fmt.Sprintf("%s/%s", group, version)
	// Remove '/' if group or version is empty
	apiver = strings.TrimPrefix(apiver, "/")
	apiver = strings.TrimSuffix(apiver, "/")
	newFlt := f.ByAPIVersion(apiver)
	if kind != "" {
		return newFlt.ByKey(yaml.KindField, kind)
	}
	return newFlt
}

// ByName adds filter by 'metadata.name' field value
func (f DocumentSelector) ByName(name string) DocumentSelector {
	if name != "" {
		return f.byPath([]string{yaml.MetadataField, yaml.NameField}, name)
	}
	return f
}

// ByNamespace adds filter by 'metadata.namespace' field value
func (f DocumentSelector) ByNamespace(ns string) DocumentSelector {
	if ns != "" {
		return f.byPath([]string{yaml.MetadataField, yaml.NamespaceField}, ns)
	}
	return f
}

// ByLabel adds filter to filter by labelSelector.
// For more details about syntax for labelSelector refer to Parse() function description from
// https://github.com/kubernetes/apimachinery/blob/master/pkg/labels/selector.go
func (f DocumentSelector) ByLabel(labelSelector string) DocumentSelector {
	if labelSelector != "" {
		f.filters = append(f.filters, &LabelFilter{MatchExpression: labelSelector})
		return f
	}
	return f
}

// Filter RNode objects
func (f DocumentSelector) Filter(items []*yaml.RNode) (result []*yaml.RNode, err error) {
	result = items
	for i := range f.filters {
		result, err = f.filters[i].Filter(result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

var _ kio.Filter = &LabelFilter{}

// LabelFilter allows to filter labels based on label Selectors
// Uses kubernetes label selector library from apimachinery package undreneath
// For syntax for MatchExpression, please refer to description of Parse() function at
// https://github.com/kubernetes/apimachinery/blob/master/pkg/labels/selector.go
type LabelFilter struct {
	MatchExpression string
}

// Filter implements RNode filter interface
func (lf *LabelFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	var output kio.ResourceNodeSlice
	selector, err := labels.Parse(lf.MatchExpression)
	if err != nil {
		return nil, err
	}

	for _, node := range input {
		nodeLabels, err := node.GetLabels()
		if err != nil {
			return nil, err
		}

		if selector.Matches(labels.Set(nodeLabels)) {
			output = append(output, node)
		}
	}

	return output, nil
}

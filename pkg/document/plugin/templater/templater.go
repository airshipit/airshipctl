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

package templater

import (
	"bytes"
	"fmt"
	"text/template"

	sprig "github.com/Masterminds/sprig/v3"
	"github.com/lucasjones/reggen"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/fs"
)

var _ kio.Filter = &plugin{}

type plugin struct {
	*airshipv1.Templater
}

// New creates new instance of the plugin
func New(obj map[string]interface{}) (kio.Filter, error) {
	cfg := &airshipv1.Templater{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj, cfg)
	if err != nil {
		return nil, err
	}
	return &plugin{
		Templater: cfg,
	}, nil
}

func (t *plugin) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	docfs := fs.NewDocumentFs()
	out := &bytes.Buffer{}
	funcMap := sprig.TxtFuncMap()
	funcMap["toUint32"] = func(i int) uint32 { return uint32(i) }
	funcMap["toYaml"] = toYaml
	funcMap["regexGen"] = regexGen
	funcMap["fileExists"] = docfs.Exists
	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(t.Template)
	if err != nil {
		return nil, err
	}
	if err = tmpl.Execute(out, t.Values); err != nil {
		return nil, err
	}

	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.ByteReader{Reader: out}},
		Outputs: []kio.Writer{&kio.PackageBuffer{}},
	}
	err = p.Execute()
	if err != nil {
		return nil, err
	}

	res, ok := p.Outputs[0].(*kio.PackageBuffer)
	if !ok {
		return nil, fmt.Errorf("output conversion error")
	}
	return append(items, res.Nodes...), nil
}

// Generate Regex
func regexGen(regex string, limit int) string {
	if limit <= 0 {
		panic("Limit cannot be less than or equal to 0")
	}
	str, err := reggen.Generate(regex, limit)
	if err != nil {
		panic(err)
	}
	return str
}

// Render input yaml as output yaml
// This function is from the Helm project:
// https://github.com/helm/helm
// Copyright The Helm Authors
func toYaml(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}

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
	"encoding/json"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"

	sprig "github.com/Masterminds/sprig/v3"

	extlib "opendev.org/airship/airshipctl/pkg/document/plugin/templater/extlib"
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

func funcMapAppend(fma, fmb template.FuncMap) template.FuncMap {
	for k, v := range fmb {
		_, ok := fma[k]
		if ok {
			panic(fmt.Errorf("trying to redefine function %s that already exists", k))
		}
		fma[k] = v
	}
	return fma
}

func (t *plugin) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	out := &bytes.Buffer{}

	funcMap := template.FuncMap{}
	funcMap = funcMapAppend(funcMap, sprig.TxtFuncMap())
	funcMap = funcMapAppend(funcMap, extlib.GenericFuncMap())

	tmpl, err := template.New("tmpl").Funcs(funcMap).Parse(t.Template)
	if err != nil {
		return nil, err
	}

	var values interface{}

	if t.Values != nil {
		if err = json.Unmarshal(t.Values.Raw, &values); err != nil {
			return nil, err
		}
	}

	if err = tmpl.Execute(out, values); err != nil {
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

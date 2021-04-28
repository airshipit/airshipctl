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
	"log"
	"os"

	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"

	sprig "github.com/Masterminds/sprig/v3"

	extlib "opendev.org/airship/airshipctl/pkg/document/plugin/templater/extlib"
)

var _ kio.Filter = &plugin{}

type plugin struct {
	*airshipv1.Templater
}

// define wrapper to call logging conditionally
func debug(x func()) {
	if os.Getenv("DEBUG_TEMPLATER") == "true" {
		x()
	}
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

func (t *plugin) loadModules(tmpl *template.Template, items []*yaml.RNode) ([]*yaml.RNode, error) {
	err := kio.Pipeline{
		Inputs: []kio.Reader{&kio.PackageBuffer{Nodes: items}},
		Filters: []kio.Filter{
			filters.GrepFilter{Path: []string{"apiVersion"}, Value: "^airshipit.org/v1alpha1$"},
			filters.GrepFilter{Path: []string{"kind"}, Value: "Templater"},
			kio.FilterFunc(func(o []*yaml.RNode) ([]*yaml.RNode, error) {
				for _, node := range o {
					templateNode, err := node.Pipe(yaml.PathGetter{Path: []string{"template"}})
					if err != nil {
						return nil, err
					}
					s := yaml.GetValue(templateNode)
					debug(func() { log.Printf("Adding module:\n%s", s) })
					_, err = tmpl.Parse(s)
					if err != nil {
						return nil, err
					}
				}
				return o, nil
			}),
		},
	}.Execute()
	if err != nil {
		return nil, err
	}

	return items, nil
}

func (t *plugin) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	out := &bytes.Buffer{}

	tmpl := template.New(t.Name)

	funcMap := template.FuncMap{}
	funcMap = funcMapAppend(funcMap, sprig.TxtFuncMap())
	funcMap = funcMapAppend(funcMap, extlib.GenericFuncMap())

	itemsFuncMap := template.FuncMap{}
	itemsFuncMap["getItems"] = func() []*yaml.RNode {
		return items
	}
	itemsFuncMap["setItems"] = func(val interface{}) error {
		newItems, err := getRNodes(val)
		if err != nil {
			return err
		}

		items = newItems
		return nil
	}
	itemsFuncMap["include"] = func(name string, data interface{}) (string, error) {
		localOut := &bytes.Buffer{}
		if err := tmpl.ExecuteTemplate(localOut, name, data); err != nil {
			return "", err
		}
		return localOut.String(), nil
	}
	funcMap = funcMapAppend(funcMap, itemsFuncMap)
	tmpl = tmpl.Funcs(funcMap)

	items, err := t.loadModules(tmpl, items)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.Parse(t.Template)
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
	debug(func() { log.Printf("Templater out is:\n%s", out.String()) })

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

func getRNodes(rnodesarr interface{}) ([]*yaml.RNode, error) {
	rnodes, ok := rnodesarr.([]*yaml.RNode)
	if ok {
		return rnodes, nil
	}

	rnodesx, ok := rnodesarr.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type %T - wanted []", rnodesarr)
	}

	rns := []*yaml.RNode{}
	for i, r := range rnodesx {
		rn, ok := r.(*yaml.RNode)
		if !ok {
			return nil, fmt.Errorf("has got element %d with unexpected type %T", i, r)
		}
		rns = append(rns, rn)
	}
	return rns, nil
}

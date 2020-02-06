/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins

import (
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/yaml"
)

// PluginRegistry map of plugin kinds to plugin instance
var PluginRegistry = make(map[string]resmap.TransformerPlugin)

// TransformerLoader airship document plugin loader. Loads external
// Kustomize plugins as builtin
type TransformerLoader struct {
	resid.ResId
}

// Config reads plugin configuration structure
func (l *TransformerLoader) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	if err := yaml.Unmarshal(c, l); err != nil {
		return err
	}
	airshipPlugin, found := PluginRegistry[l.Kind]
	if !found {
		return ErrUnknownPlugin{Kind: l.Kind}
	}
	return airshipPlugin.Config(ldr, rf, c)
}

// Transform executes Transform method of an external plugin
func (l *TransformerLoader) Transform(m resmap.ResMap) error {
	airshipPlugin, found := PluginRegistry[l.Kind]
	if !found {
		return ErrUnknownPlugin{Kind: l.Kind}
	}
	return airshipPlugin.Transform(m)
}

// NewTransformerLoader returns plugin loader instance
func NewTransformerLoader() resmap.TransformerPlugin {
	return &TransformerLoader{}
}

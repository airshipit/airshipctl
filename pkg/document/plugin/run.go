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

package plugin

import (
	"io"

	"sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/document/plugin/replacement"
	"opendev.org/airship/airshipctl/pkg/document/plugin/templater"
	"opendev.org/airship/airshipctl/pkg/document/plugin/types"
)

// DefaultPlugins returns map with plugin factories
func DefaultPlugins() (map[schema.GroupVersionKind]types.Factory, error) {
	registry := make(map[schema.GroupVersionKind]types.Factory)
	if err := replacement.RegisterPlugin(registry); err != nil {
		return nil, err
	}
	if err := templater.RegisterPlugin(registry); err != nil {
		return nil, err
	}
	return registry, nil
}

// ConfigureAndRun executes particular plugin based on group, version, kind
// which have been specified in configuration file. Config file should be
// supplied as a first element of args slice
func ConfigureAndRun(pluginCfg []byte, in io.Reader, out io.Writer) error {
	rawCfg := make(map[string]interface{})
	if err := yaml.Unmarshal(pluginCfg, &rawCfg); err != nil {
		return err
	}
	uCfg := &unstructured.Unstructured{}
	uCfg.SetUnstructuredContent(rawCfg)
	registry, err := DefaultPlugins()
	if err != nil {
		return err
	}
	pluginFactory, ok := registry[uCfg.GroupVersionKind()]
	if !ok {
		return ErrPluginNotFound{PluginID: uCfg.GroupVersionKind()}
	}

	plugin, err := pluginFactory(rawCfg)
	if err != nil {
		return err
	}
	return plugin.Run(in, out)
}

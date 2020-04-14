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

	"opendev.org/airship/airshipctl/pkg/document/plugin/types"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Registry contains factory functions for the available plugins
var Registry = make(map[schema.GroupVersionKind]types.Factory)

// ConfigureAndRun executes particular plugin based on group, version, kind
// which have been specified in configuration file. Config file should be
// supplied as a first element of args slice
func ConfigureAndRun(settings *environment.AirshipCTLSettings, pluginCfg []byte, in io.Reader, out io.Writer) error {
	var cfg unstructured.Unstructured
	if err := yaml.Unmarshal(pluginCfg, &cfg); err != nil {
		return err
	}
	pluginFactory, ok := Registry[cfg.GroupVersionKind()]
	if !ok {
		return ErrPluginNotFound{PluginID: cfg.GroupVersionKind()}
	}

	plugin, err := pluginFactory(settings, pluginCfg)
	if err != nil {
		return err
	}
	return plugin.Run(in, out)
}

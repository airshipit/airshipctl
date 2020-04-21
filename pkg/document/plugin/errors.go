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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ErrPluginNotFound is returned if a plugin was not found in the plugin
// registry
type ErrPluginNotFound struct {
	//PluginID group, version and kind plugin identifier
	PluginID schema.GroupVersionKind
}

func (e ErrPluginNotFound) Error() string {
	return fmt.Sprintf("plugin identified by %s was not found", e.PluginID.String())
}

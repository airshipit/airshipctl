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

package implementations

import (
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	"opendev.org/airship/airshipctl/pkg/log"
)

var _ repository.ComponentsClient = &ComponentsClient{}

// ComponentsClient override Get() method to return same components,
// but in our implementation we skip variable substitution.
type ComponentsClient struct {
	client               repository.ComponentsClient
	providerType         string
	providerName         string
	variableSubstitution bool
}

// Get returns the components from a repository but without variable substitution
func (cc *ComponentsClient) Get(options repository.ComponentsOptions) (repository.Components, error) {
	// Invert variable substitution, so that by default clusterctl will not substitute variables
	options.SkipVariables = !cc.variableSubstitution
	log.Printf("Getting airshipctl provider components, skipping variable substitution: %t.\n"+
		"Provider type: %s, name: %s\n", options.SkipVariables, cc.providerType, cc.providerName)
	return cc.client.Get(options)
}

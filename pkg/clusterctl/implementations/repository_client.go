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
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	"opendev.org/airship/airshipctl/pkg/log"
)

var _ config.Provider = &RepositoryClient{}

// RepositoryClient override Components() method to return same components client,
// but in our implementation we skip variable substitution.
type RepositoryClient struct {
	repository.Client
	ProviderType string
	ProviderName string
}

// Components provide access to YAML file for creating provider components.
func (rc *RepositoryClient) Components() repository.ComponentsClient {
	log.Debugf("Setting up airshipctl provider Components client\n"+
		"Provider type: %s, name: %s\n", rc.ProviderType, rc.ProviderName)
	return &ComponentsClient{
		client:       rc.Client.Components(),
		providerName: rc.ProviderName,
		providerType: rc.ProviderType,
	}
}

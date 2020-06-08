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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

func TestProvider(t *testing.T) {
	cctl := &Clusterctl{
		Providers: []*Provider{
			{
				Name:                   "kubeadm",
				URL:                    "/home/providers/kubeadm/v0.3.5/components.yaml",
				Type:                   "BootstrapProvider",
				IsClusterctlRepository: true,
			},
		},
	}
	tests := []struct {
		name             string
		getName          string
		getType          string
		expectedProvider *Provider
	}{
		{
			name:    "repo options exist",
			getName: "kubeadm",
			getType: "BootstrapProvider",
			expectedProvider: &Provider{
				Name:                   "kubeadm",
				URL:                    "/home/providers/kubeadm/v0.3.5/components.yaml",
				Type:                   "BootstrapProvider",
				IsClusterctlRepository: true,
			},
		},
		{
			name:             "repo name does not exist",
			getName:          "does not exist",
			getType:          "BootstrapProvider",
			expectedProvider: nil,
		},
		{
			name:             "type does not exist",
			getName:          "kubeadm",
			getType:          "does not exist",
			expectedProvider: nil,
		},
	}

	for _, tt := range tests {
		getName := tt.getName
		getType := tt.getType
		expectedProvider := tt.expectedProvider
		t.Run(tt.name, func(t *testing.T) {
			actualProvider := cctl.Provider(getName, clusterctlv1.ProviderType(getType))
			assert.Equal(t, expectedProvider, actualProvider)
		})
	}
}

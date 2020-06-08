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

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	"sigs.k8s.io/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/clusterctl/api/v1alpha1"
)

var (
	testConfig = `apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: clusterctl-v1
init-options: {}
providers:
- name: "aws"
  type: "InfrastructureProvider"
  url: "/manifests/capi/infra/aws/v0.3.0"
  clusterctl-repository: true
- name: "custom-infra"
  type: "InfrastructureProvider"
  url: "/manifests/capi/custom-infra/aws/v0.3.0"
  clusterctl-repository: true
- name: "custom-airship-infra"
  type: "InfrastructureProvider"
  versions:
    v0.3.1: functions/capi/infrastructure/v0.3.1
    v0.3.2: functions/capi/infrastructure/v0.3.2`
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name            string
		conf            *airshipv1.Clusterctl
		presentProvider string
		presentType     string
		expectedURL     string
	}{
		{
			name:            "clusterctl single repo",
			presentProvider: "kubeadm",
			presentType:     "BootstrapProvider",
			expectedURL:     "/home/providers/kubeadm/v0.3.5/components.yaml",
			conf: &airshipv1.Clusterctl{

				Providers: []*airshipv1.Provider{
					{
						Name:                   "kubeadm",
						URL:                    "/home/providers/kubeadm/v0.3.5/components.yaml",
						Type:                   "BootstrapProvider",
						IsClusterctlRepository: true,
					},
				},
			},
		},
		{
			name:            "multiple repos with airship",
			presentProvider: "airship-repo",
			presentType:     "InfrastructureProvider",
			expectedURL:     testDataDir,
			conf: &airshipv1.Clusterctl{

				Providers: []*airshipv1.Provider{
					{
						Name:                   "airship-repo",
						URL:                    "/home/providers/my-repo/v0.3.5/components.yaml",
						Type:                   "InfrastructureProvider",
						IsClusterctlRepository: false,
						Versions: map[string]string{
							"v0.3.1": "some-path",
						},
					},
					{
						Name:                   "kubeadm",
						URL:                    "/home/providers/kubeadm/v0.3.5/components.yaml",
						Type:                   "BootstrapProvider",
						IsClusterctlRepository: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		conf := tt.conf
		url := tt.expectedURL
		provName := tt.presentProvider
		provType := tt.presentType
		t.Run(tt.name, func(t *testing.T) {
			got, err := newConfig(conf, testDataDir)
			require.NoError(t, err)
			providerClient := got.Providers()
			provider, err := providerClient.Get(provName, clusterctlv1.ProviderType(provType))
			require.NoError(t, err)
			assert.Equal(t, url, provider.URL())
		})
	}
}

func TestNewClientEmptyOptions(t *testing.T) {
	c := &airshipv1.Clusterctl{}
	client, err := NewClient("", true, c)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestNewClient(t *testing.T) {
	c := &airshipv1.Clusterctl{}
	err := yaml.Unmarshal([]byte(testConfig), c)
	require.NoError(t, err)

	client, err := NewClient("", true, c)
	require.NoError(t, err)
	require.NotNil(t, client)
}

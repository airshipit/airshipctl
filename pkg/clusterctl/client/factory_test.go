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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlclient "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/yamlprocessor"
	"sigs.k8s.io/yaml"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

const (
	testDataDir = "testdata"
)

var (
	testConfigFactory = `apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: clusterctl-v1
init-options: {}
providers:
  - name: "aws"
    type: "InfrastructureProvider"
    url: "/manifests/capi/infra/infrastructure-aws/v0.3.0/components.yaml"
    clusterctl-repository: true
  - name: "custom-infra"
    type: "InfrastructureProvider"
    url: "/manifests/capi/infra/infrastructure-custom-infra/v0.3.0/components.yaml"
    clusterctl-repository: true
  - name: "custom-airship-infra"
    type: "InfrastructureProvider"
    versions:
      v0.3.1: functions/capi/infrastructure/v0.3.1
      v0.3.2: functions/capi/infrastructure/v0.3.2`
)

func testOptions(t *testing.T, input string) *airshipv1.Clusterctl {
	t.Helper()
	o := &airshipv1.Clusterctl{}
	err := yaml.Unmarshal([]byte(input), o)
	require.NoError(t, err)
	return o
}

func testNewConfig(t *testing.T, o *airshipv1.Clusterctl) clusterctlconfig.Client {
	t.Helper()
	configClient, err := newConfig(o, testDataDir)
	require.NoError(t, err)
	require.NotNil(t, configClient)
	return configClient
}

// TestFactory checks if airship repository interface is selected for providers that are not
// of airship type, and that this interface methods return correct components
func TestFactory(t *testing.T) {
	o := testOptions(t, testConfigFactory)
	configClient := testNewConfig(t, o)
	factory := RepositoryFactory{
		Options:      o,
		ConfigClient: configClient,
	}
	repoFactory := factory.ClientRepositoryFactory()
	require.NotNil(t, repoFactory)
	pclient := configClient.Providers()
	require.NotNil(t, pclient)
	tests := []struct {
		name              string
		expectedVersions  []string
		useVersion        string
		useName           string
		useType           string
		expectErr         bool
		expectedNamespace string
	}{
		{
			name:              "custom airship v1",
			expectedVersions:  []string{"v0.3.1", "v0.3.2"},
			useVersion:        "v0.3.1",
			useName:           "custom-airship-infra",
			useType:           "InfrastructureProvider",
			expectErr:         false,
			expectedNamespace: "version-one",
		},
		{
			name:              "custom airship v2",
			expectedVersions:  []string{"v0.3.1", "v0.3.2"},
			useVersion:        "v0.3.2",
			useName:           "custom-airship-infra",
			useType:           "InfrastructureProvider",
			expectErr:         false,
			expectedNamespace: "version-two",
		},
	}
	for _, tt := range tests {
		expectedVersions := tt.expectedVersions
		useVersion := tt.useVersion
		expectErr := tt.expectErr
		useName := tt.useName
		useType := tt.useType
		expectedNamespace := tt.expectedNamespace
		t.Run(tt.name, func(t *testing.T) {
			provider, err := pclient.Get(useName, clusterctlv1.ProviderType(useType))
			require.NoError(t, err)
			require.NotNil(t, provider)
			repo, err := repoFactory(clusterctlclient.RepositoryClientFactoryInput{
				Provider:  provider,
				Processor: yamlprocessor.NewSimpleProcessor(),
			})
			require.NoError(t, err)
			require.NotNil(t, repo)
			versions, err := repo.GetVersions()
			require.NoError(t, err)
			sort.Strings(expectedVersions)
			sort.Strings(versions)
			assert.Equal(t, testDataDir, repo.URL())
			assert.Equal(t, expectedVersions, versions)
			components := repo.Components()
			require.NotNil(t, components)
			// namespaces are left blank, since namespace is provided in the document set
			component, err := components.Get(repository.ComponentsOptions{
				Version: useVersion,
			})
			require.NoError(t, err)
			require.NotNil(t, component)

			b, err := component.Yaml()
			if expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				actualNamespace := &v1.Namespace{}
				err = yaml.Unmarshal(b, actualNamespace)
				require.NoError(t, err)
				assert.Equal(t, expectedNamespace, actualNamespace.GetName())
			}
		})
	}
}

func TestClientRepositoryFactory(t *testing.T) {
	o := testOptions(t, testConfigFactory)
	configClient := testNewConfig(t, o)
	factory := RepositoryFactory{
		Options:      o,
		ConfigClient: configClient,
	}
	clusterclientFactory := factory.ClusterClientFactory()
	clusterClient, err := clusterclientFactory(clusterctlclient.ClusterClientFactoryInput{
		Kubeconfig: clusterctlclient.Kubeconfig{
			Path:    "testdata/kubeconfig.yaml",
			Context: ""},
		Processor: yamlprocessor.NewSimpleProcessor(),
	})
	assert.NoError(t, err)
	assert.NotNil(t, clusterClient)
}

func TestRepoFactoryFunction(t *testing.T) {
	o := testOptions(t, testConfigFactory)
	configClient := testNewConfig(t, o)

	factory := RepositoryFactory{
		Options:      o,
		ConfigClient: configClient,
	}
	pclient := configClient.Providers()
	require.NotNil(t, pclient)
	provider, err := pclient.Get("custom-airship-infra", "InfrastructureProvider")
	require.NoError(t, err)
	repoClient, err := factory.repoFactory(clusterctlclient.RepositoryClientFactoryInput{
		Provider:  provider,
		Processor: yamlprocessor.NewSimpleProcessor(),
	})
	require.NoError(t, err)
	require.NotNil(t, repoClient)
	versions, err := repoClient.GetVersions()
	expectedVersions := []string{"v0.3.1", "v0.3.2"}
	sort.Strings(versions)
	sort.Strings(expectedVersions)
	require.NoError(t, err)
	assert.Equal(t, expectedVersions, versions)
}

func TestClusterctlRepoFactoryFunction(t *testing.T) {
	o := testOptions(t, testConfigFactory)
	configClient := testNewConfig(t, o)
	factory := RepositoryFactory{
		Options:      o,
		ConfigClient: configClient,
	}
	pclient := configClient.Providers()
	provider, err := pclient.Get("aws", "InfrastructureProvider")
	require.NoError(t, err)
	repoClient, err := factory.repoFactory(clusterctlclient.RepositoryClientFactoryInput{
		Provider:  provider,
		Processor: yamlprocessor.NewSimpleProcessor(),
	})
	require.NoError(t, err)
	require.NotNil(t, repoClient)
}

// Test error cases
func TestRepositoryFactoryErrors(t *testing.T) {
	// set one default provider clusterctl is properly initialized
	defProv := &airshipv1.Provider{
		Name: "aws",
		Type: "InfrastructureProvider",
		Versions: map[string]string{
			"v0.3.3": testConfig + "/functions/capi/v0.3.3",
		},
	}
	o := &airshipv1.Clusterctl{
		Providers: []*airshipv1.Provider{defProv},
	}
	configClient := testNewConfig(t, o)
	require.NotNil(t, configClient)
	factory := RepositoryFactory{
		Options:      o,
		ConfigClient: configClient,
	}
	rf := factory.ClientRepositoryFactory()
	require.NotNil(t, rf)
	pclient := configClient.Providers()
	require.NotNil(t, pclient)
	// save provider so then we can run tests against it, while modifying original airship clustetrctl conf
	provider, err := pclient.Get("aws", "InfrastructureProvider")
	require.NoError(t, err)
	require.NotNil(t, provider)
	tests := []struct {
		name     string
		airProvs []*airshipv1.Provider
	}{
		{
			name:     "providers are nil",
			airProvs: nil,
		},
		{
			name: "versions are nil",
			airProvs: []*airshipv1.Provider{
				{
					Name:     "aws",
					Type:     "InfrastructureProvider",
					Versions: nil,
				},
			},
		},
		{
			name: "versions can't be parsed",
			airProvs: []*airshipv1.Provider{
				{
					Name: "aws",
					Type: "InfrastructureProvider",
					Versions: map[string]string{
						"can't parse version": "wrong path",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		airProvs := tt.airProvs
		t.Run(tt.name, func(t *testing.T) {
			// set airship providers so it does not correspond to clusterctl provider
			o.Providers = airProvs
			crc, err := factory.repoFactory(clusterctlclient.RepositoryClientFactoryInput{
				Provider:  provider,
				Processor: yamlprocessor.NewSimpleProcessor(),
			})
			// expect error since we have mismatch of airship providers vs clusterctl providers
			require.Nil(t, crc)
			assert.Error(t, err)
		})
	}
}

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

func TestRepositoryClient(t *testing.T) {
	providerName := "metal3"
	providerType := "InfrastructureProvider"
	// this version contains a variable that is suppose to be substituted by clusterctl
	// and we will test if the variable is found and not substituted
	versions := map[string]string{
		"v0.2.3": "functions/4",
	}
	options := &airshipv1.Clusterctl{
		Providers: []*airshipv1.Provider{
			{
				Name:     providerName,
				Type:     providerType,
				URL:      "/dummy/path/v0.3.2/components.yaml",
				Versions: versions,
			},
		},
	}
	// create instance of airship reader interface implementation for clusterctl and inject it
	reader, err := NewAirshipReader(options)
	require.NoError(t, err)
	require.NotNil(t, reader)
	optionReader := clusterctlconfig.InjectReader(reader)
	require.NotNil(t, optionReader)
	configClient, err := clusterctlconfig.New("", optionReader)
	require.NoError(t, err)
	require.NotNil(t, configClient)
	// get the provider from configuration client, in which we injected our reader
	provider, err := configClient.Providers().Get(providerName, clusterctlv1.ProviderType(providerType))
	require.NoError(t, err)
	require.NotNil(t, provider)
	// Create instance of airship repository interface implementation for clusterctl
	repo, err := NewRepository("testdata", versions)
	require.NoError(t, err)
	require.NotNil(t, repo)
	// Inject the repository in repository client
	optionsRepo := repository.InjectRepository(repo)
	repoClient, err := repository.New(provider, configClient, optionsRepo)
	require.NoError(t, err)
	require.NotNil(t, repoClient)
	// create airship implementation of clusterctl repository client
	airRepoClient := RepositoryClient{
		Client: repoClient,
	}
	// get the components of the repository with empty options, all defaults should work
	c, err := airRepoClient.Components().Get(repository.ComponentsOptions{})
	require.NoError(t, err)
	// No errors must be returned since there is are no variables that need to be substituted
	assert.NotNil(t, c)
	// Make sure that target namespace is the same as defined by repository implementation bundle
	assert.Equal(t, "newnamespace", c.TargetNamespace())
	// Make sure that variables for substitution are actually found
	require.Len(t, c.Variables(), 1)
	// make sure that variable name is correct
	assert.Equal(t, "PROVISIONING_IP", c.Variables()[0])
}

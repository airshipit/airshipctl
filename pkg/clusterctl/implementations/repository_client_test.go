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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctlconfig "sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/repository"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

func TestRepositoryClient(t *testing.T) {
	airRepoClient := testRepoClient(testRepoOpts{
		kustRoot:       "functions/4",
		envVars:        false,
		additionalVars: map[string]string{},
	}, t)
	// get the components of the repository with empty options, all defaults should work
	// SkipVariables is to true, to make sure that it is ignored in this implementation, and instead
	// taken from airship clusterctl provider option, which disables var substitution by default
	c, err := airRepoClient.Components().Get(repository.ComponentsOptions{SkipVariables: true})
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

func TestMissingVariableRepoClient(t *testing.T) {
	airRepoClient := testRepoClient(testRepoOpts{
		kustRoot:        "functions/5",
		envVars:         true,
		additionalVars:  map[string]string{},
		varSubstitution: true,
	}, t)
	envVars := map[string]string{
		"AZURE_SUBSCRIPTION_ID_B64": "c29tZS1iYXNlNjQtSUQtdGV4dAo=",
		"AZURE_TENANT_ID_B64":       "c29tZS1iYXNlNjQtVEVOQU5ULUlELXRleHQK",
		"AZURE_CLIENT_ID_B64":       "c29tZS1iYXNlNjQtQ0xJRU5ULUlELXRleHQK",
	}
	for key, val := range envVars {
		os.Setenv(key, val)
		defer os.Unsetenv(key)
	}
	c, err := airRepoClient.Components().Get(repository.ComponentsOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `value for variables [AZURE_CLIENT_SECRET_B64] is not set`)
	assert.Nil(t, c)
}

func TestEnvVariableSubstiutionRepoClient(t *testing.T) {
	airRepoClient := testRepoClient(testRepoOpts{
		kustRoot:        "functions/5",
		envVars:         true,
		additionalVars:  map[string]string{},
		varSubstitution: true,
	}, t)
	envVars := map[string]string{
		"AZURE_SUBSCRIPTION_ID_B64": "c29tZS1iYXNlNjQtSUQtdGV4dAo=",
		"AZURE_TENANT_ID_B64":       "c29tZS1iYXNlNjQtVEVOQU5ULUlELXRleHQK",
		"AZURE_CLIENT_ID_B64":       "c29tZS1iYXNlNjQtQ0xJRU5ULUlELXRleHQK",
		"AZURE_CLIENT_SECRET_B64":   "c29tZS1iYXNlNjQtQ0xJRU5ULVNFQ1JFVC10ZXh0Cg==",
	}
	for key, val := range envVars {
		os.Setenv(key, val)
		defer os.Unsetenv(key)
	}
	c, err := airRepoClient.Components().Get(repository.ComponentsOptions{})
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.Len(t, c.Variables(), len(dataKeyMapping()))
	// find secret containing env variables
	for _, obj := range c.InstanceObjs() {
		if obj.GetKind() == "Secret" {
			cm := &v1.ConfigMap{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), cm)
			require.NoError(t, err)
			for key, expectedVal := range envVars {
				dataKey, exists := dataKeyMapping()[key]
				require.True(t, exists)
				actualVal, exists := cm.Data[dataKey]
				require.True(t, exists)
				assert.Equal(t, expectedVal, actualVal)
			}
		}
	}
}

// This test covers a case, where we want some variables to be substituted and some
// are not. Clusterctl behavior doesn't allow to skip variable substitution completely
// instead if SkipVariables is set to True, it will not throw errors if these variables
// are not set in config reader.
func TestAdditionalVariableSubstiutionRepoClient(t *testing.T) {
	vars := map[string]string{
		"AZURE_SUBSCRIPTION_ID_B64": "c29tZS1iYXNlNjQtSUQtdGV4dAo=",
		"AZURE_TENANT_ID_B64":       "c29tZS1iYXNlNjQtVEVOQU5ULUlELXRleHQK",
		"AZURE_CLIENT_ID_B64":       "c29tZS1iYXNlNjQtQ0xJRU5ULUlELXRleHQK",
	}
	notSubstitutedVars := map[string]string{
		"AZURE_CLIENT_SECRET_B64": "${AZURE_CLIENT_SECRET_B64}",
	}
	airRepoClient := testRepoClient(testRepoOpts{
		kustRoot:       "functions/5",
		envVars:        false,
		additionalVars: vars,
		// set to false so errors are not thrown when AZURE_CLIENT_SECRET_B64 is not found
		varSubstitution: false,
	}, t)
	c, err := airRepoClient.Components().Get(repository.ComponentsOptions{})
	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.Len(t, c.Variables(), len(dataKeyMapping()))
	// find secret containing env variables
	for _, obj := range c.InstanceObjs() {
		if obj.GetKind() == "Secret" {
			// merge two maps
			mergedVars := map[string]string{}
			for k, v := range vars {
				mergedVars[k] = v
			}
			for k, v := range notSubstitutedVars {
				mergedVars[k] = v
			}
			cm := &v1.ConfigMap{}
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), cm)
			require.NoError(t, err)
			for key, expectedVal := range vars {
				dataKey, exists := dataKeyMapping()[key]
				require.True(t, exists)
				actualVal, exists := cm.Data[dataKey]
				require.True(t, exists)
				assert.Equal(t, expectedVal, actualVal)
			}
		}
	}
}

type testRepoOpts struct {
	kustRoot        string
	envVars         bool
	additionalVars  map[string]string
	varSubstitution bool
}

func testRepoClient(opts testRepoOpts, t *testing.T) repository.Client {
	providerName := "metal3"
	providerType := "InfrastructureProvider"
	// this version contains a variable that is suppose to be substituted by clusterctl
	// and we will test if the variable is found and not substituted
	versions := map[string]string{
		"v0.2.3": opts.kustRoot,
	}
	cctl := &airshipv1.Clusterctl{
		AdditionalComponentVariables: opts.additionalVars,
		EnvVars:                      opts.envVars,
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
	reader, err := NewAirshipReader(cctl)
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
	return &RepositoryClient{
		ProviderName:         providerName,
		ProviderType:         providerType,
		Client:               repoClient,
		VariableSubstitution: opts.varSubstitution,
	}
}

func dataKeyMapping() map[string]string {
	return map[string]string{
		"AZURE_SUBSCRIPTION_ID_B64": "subscription-id",
		"AZURE_TENANT_ID_B64":       "tenant-id",
		"AZURE_CLIENT_ID_B64":       "client-id",
		"AZURE_CLIENT_SECRET_B64":   "client-secret",
	}
}

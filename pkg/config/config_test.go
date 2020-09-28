/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	stringDelta        = "_changed"
	currentContextName = "def_ephemeral"
	defaultString      = "default"
)

func TestString(t *testing.T) {
	fSys := testutil.SetupTestFs(t, "testdata")

	tests := []struct {
		name     string
		stringer fmt.Stringer
	}{
		{
			name:     "config",
			stringer: testutil.DummyConfig(),
		},
		{
			name:     "context",
			stringer: testutil.DummyContext(),
		},
		{
			name:     "manifest",
			stringer: testutil.DummyManifest(),
		},
		{
			name:     "repository",
			stringer: testutil.DummyRepository(),
		},
		{
			name:     "repo-auth",
			stringer: testutil.DummyRepoAuth(),
		},
		{
			name:     "repo-checkout",
			stringer: testutil.DummyRepoCheckout(),
		},
		{
			name:     "managementconfiguration",
			stringer: testutil.DummyManagementConfiguration(),
		},
		{
			name:     "encryption-config",
			stringer: testutil.DummyEncryptionConfig(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			filename := fmt.Sprintf("/%s-string.yaml", tt.name)
			data, err := fSys.ReadFile(filename)
			require.NoError(t, err)

			assert.Equal(t, string(data), tt.stringer.String())
		})
	}
}

func TestLoadConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	assert.Len(t, conf.Contexts, 4)
}

func TestPersistConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.SetLoadedConfigPath(conf.LoadedConfigPath() + ".new")
	conf.SetKubeConfigPath(conf.KubeConfigPath() + ".new")

	err := conf.PersistConfig(true)
	require.NoError(t, err)

	// Check that the files were created
	assert.FileExists(t, conf.LoadedConfigPath())
	assert.FileExists(t, conf.KubeConfigPath())
	// Check that the invalid name was changed to a valid one
	assert.Contains(t, conf.KubeConfig().Clusters, "def_ephemeral")
}

func TestEnsureComplete(t *testing.T) {
	// This test is intentionally verbose. Since a user of EnsureComplete
	// does not need to know about the order of validation, each test
	// object passed into EnsureComplete should have exactly one issue, and
	// be otherwise valid
	tests := []struct {
		name        string
		config      config.Config
		expectedErr error
	}{
		{
			name: "no contexts defined",
			config: config.Config{
				Contexts:       map[string]*config.Context{},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one Context needs to be defined"},
		},
		{
			name: "no manifests defined",
			config: config.Config{
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one Manifest needs to be defined"},
		},
		{
			name: "current context not defined",
			config: config.Config{
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context is not defined"},
		},
		{
			name: "no context for current context",
			config: config.Config{
				Contexts:       map[string]*config.Context{"DIFFERENT_CONTEXT": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Context"},
		},
		{
			name: "no manifest for current context",
			config: config.Config{
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"DIFFERENT_MANIFEST": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Manifest"},
		},
		{
			name: "complete config",
			config: config.Config{
				EncryptionConfigs: map[string]*config.EncryptionConfig{"testEncryptionConfig": {}},
				Contexts:          map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:         map[string]*config.Manifest{"testManifest": {}},
				CurrentContext:    "testContext",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			actualErr := tt.config.EnsureComplete()
			assert.Equal(subTest, tt.expectedErr, actualErr)
		})
	}
}

func TestCurrentContextManagementConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	managementConfig, err := conf.CurrentContextManagementConfig()
	require.Error(t, err)
	assert.Nil(t, managementConfig)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].ManagementConfiguration = defaultString
	conf.Contexts[currentContextName].Manifest = defaultString

	managementConfig, err = conf.CurrentContextManagementConfig()
	require.NoError(t, err)
	assert.Equal(t, conf.ManagementConfiguration[defaultString], managementConfig)
}

func TestPurge(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	// Store it
	err := conf.PersistConfig(true)
	assert.NoErrorf(t, err, "Unable to persist configuration expected at %v", conf.LoadedConfigPath())

	// Verify that the file is there
	_, err = os.Stat(conf.LoadedConfigPath())
	assert.Falsef(t, os.IsNotExist(err), "Test config was not persisted at %v, cannot validate Purge",
		conf.LoadedConfigPath())

	// Delete it
	err = conf.Purge()
	assert.NoErrorf(t, err, "Unable to Purge file at %v", conf.LoadedConfigPath())

	// Verify its gone
	_, err = os.Stat(conf.LoadedConfigPath())
	assert.Falsef(t, os.IsExist(err), "Purge failed to remove file at %v", conf.LoadedConfigPath())
}

func TestSetLoadedConfigPath(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	testPath := "/tmp/loadedconfig"

	assert.NotEqual(t, testPath, conf.LoadedConfigPath())
	conf.SetLoadedConfigPath(testPath)
	assert.Equal(t, testPath, conf.LoadedConfigPath())
}

func TestSetKubeConfigPath(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	testPath := "/tmp/kubeconfig"

	assert.NotEqual(t, testPath, conf.KubeConfigPath())
	conf.SetKubeConfigPath(testPath)
	assert.Equal(t, testPath, conf.KubeConfigPath())
}

func TestGetContexts(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	contexts := conf.GetContexts()
	assert.Len(t, contexts, 4)
}

func TestGetContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	context, err := conf.GetContext("def_ephemeral")
	require.NoError(t, err)

	// Test Positives
	assert.EqualValues(t, context.NameInKubeconf, "def_ephemeral")

	// Test Wrong Cluster
	_, err = conf.GetContext("unknown")
	assert.Error(t, err)
}

func TestAddContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyContextOptions()
	context := conf.AddContext(co)
	assert.EqualValues(t, conf.Contexts[co.Name], context)
}

func TestModifyContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyContextOptions()
	context := conf.AddContext(co)

	co.Manifest += stringDelta
	conf.ModifyContext(context, co)
	assert.EqualValues(t, conf.Contexts[co.Name].Manifest, co.Manifest)
	assert.EqualValues(t, conf.Contexts[co.Name], context)
}

func TestGetCurrentContext(t *testing.T) {
	t.Run("getCurrentContext", func(t *testing.T) {
		conf, cleanup := testutil.InitConfig(t)
		defer cleanup(t)

		context, err := conf.GetCurrentContext()
		require.Error(t, err)
		assert.Nil(t, context)

		conf.CurrentContext = currentContextName
		conf.Contexts[currentContextName].Manifest = defaultString

		context, err = conf.GetCurrentContext()
		require.NoError(t, err)
		assert.Equal(t, conf.Contexts[currentContextName], context)
	})
}

func TestCurrentContextManifest(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	manifest, err := conf.CurrentContextManifest()
	require.Error(t, err)
	assert.Nil(t, manifest)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	manifest, err = conf.CurrentContextManifest()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString], manifest)
}

func TestCurrentTargetPath(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	manifest, err := conf.CurrentContextManifest()
	require.Error(t, err)
	assert.Nil(t, manifest)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	targetPath, err := conf.CurrentContextTargetPath()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString].TargetPath, targetPath)
}

func TestCurrentContextEntryPoint(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	entryPoint, err := conf.CurrentContextEntryPoint(defaultString)
	require.Error(t, err)
	assert.Equal(t, "", entryPoint)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	entryPoint, err = conf.CurrentContextEntryPoint(defaultString)
	assert.Equal(t, config.ErrMissingPhaseDocument{PhaseName: defaultString}, err)
	assert.Nil(t, nil, entryPoint)
}

func TestCurrentContextClusterType(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	expectedClusterType := "ephemeral"

	clusterTypeEmpty, err := conf.CurrentContextClusterType()
	require.Error(t, err)
	assert.Equal(t, "", clusterTypeEmpty)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	actualClusterType, err := conf.CurrentContextClusterType()
	require.NoError(t, err)
	assert.Equal(t, expectedClusterType, actualClusterType)
}

func TestCurrentContextClusterName(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	expectedClusterName := "def"

	clusterNameEmpty, err := conf.CurrentContextClusterName()
	require.Error(t, err)
	assert.Equal(t, "", clusterNameEmpty)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	actualClusterName, err := conf.CurrentContextClusterName()
	require.NoError(t, err)
	assert.Equal(t, expectedClusterName, actualClusterName)
}

func TestCurrentContextManifestMetadata(t *testing.T) {
	expectedMeta := &config.Metadata{
		Inventory: &config.InventoryMeta{
			Path: "manifests/site/inventory",
		},
		PhaseMeta: &config.PhaseMeta{
			Path: "manifests/site/phases",
		},
	}
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)
	tests := []struct {
		name           string
		metaPath       string
		currentContext string
		expectErr      bool
		errorChecker   func(error) bool
		meta           *config.Metadata
	}{
		{
			name:           "default metadata",
			metaPath:       "metadata.yaml",
			expectErr:      false,
			currentContext: "testContext",
			meta: &config.Metadata{
				Inventory: &config.InventoryMeta{
					Path: "manifests/site/inventory",
				},
				PhaseMeta: &config.PhaseMeta{
					Path: "manifests/site/phases",
				},
			},
		},
		{
			name:           "no such file or directory",
			metaPath:       "does not exist",
			currentContext: "testContext",
			expectErr:      true,
			errorChecker:   os.IsNotExist,
		},
		{
			name:           "missing context",
			currentContext: "doesn't exist",
			expectErr:      true,
			errorChecker: func(err error) bool {
				return strings.Contains(err.Error(), "Missing configuration")
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			context := &config.Context{
				Manifest: "testManifest",
			}
			manifest := &config.Manifest{
				MetadataPath: tt.metaPath,
				TargetPath:   "testdata",
			}
			conf.Manifests = map[string]*config.Manifest{
				"testManifest": manifest,
			}
			conf.Contexts = map[string]*config.Context{
				"testContext": context,
			}
			conf.CurrentContext = tt.currentContext
			meta, err := conf.CurrentContextManifestMetadata()
			if tt.expectErr {
				t.Logf("error is %v", err)
				require.Error(t, err)
				require.NotNil(t, tt.errorChecker)
				assert.True(t, tt.errorChecker(err))
			} else {
				require.NoError(t, err)
				require.NotNil(t, meta)
				assert.Equal(t, expectedMeta, meta)
			}
		})
	}
}

func TestNewClusterComplexNameFromKubeClusterName(t *testing.T) {
	tests := []struct {
		name         string
		inputName    string
		expectedName string
		expectedType string
	}{
		{
			name:         "single-word",
			inputName:    "myCluster",
			expectedName: "myCluster",
			expectedType: config.AirshipDefaultClusterType,
		},
		{
			name:         "multi-word",
			inputName:    "myCluster_two",
			expectedName: "myCluster_two",
			expectedType: config.AirshipDefaultClusterType,
		},
		{
			name:         "cluster-appended",
			inputName:    "myCluster_ephemeral",
			expectedName: "myCluster",
			expectedType: config.Ephemeral,
		},
		{
			name:         "multi-word-cluster-appended",
			inputName:    "myCluster_two_ephemeral",
			expectedName: "myCluster_two",
			expectedType: config.Ephemeral,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			complexName := config.NewClusterComplexNameFromKubeClusterName(tt.inputName)
			assert.Equal(t, tt.expectedName, complexName.Name)
			assert.Equal(t, tt.expectedType, complexName.Type)
		})
	}
}

func TestManagementConfigurationByName(t *testing.T) {
	conf, cleanupConfig := testutil.InitConfig(t)
	defer cleanupConfig(t)

	mgmtCfg, err := conf.GetManagementConfiguration(config.AirshipDefaultContext)
	require.NoError(t, err)
	assert.Equal(t, conf.ManagementConfiguration[config.AirshipDefaultContext], mgmtCfg)
}

func TestManagementConfigurationByNameDoesNotExist(t *testing.T) {
	conf, cleanupConfig := testutil.InitConfig(t)
	defer cleanupConfig(t)

	_, err := conf.GetManagementConfiguration(fmt.Sprintf("%s-test", config.AirshipDefaultContext))
	assert.Error(t, err)
}

func TestGetManifests(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	manifests := conf.GetManifests()
	require.NotNil(t, manifests)

	assert.EqualValues(t, manifests[0].PrimaryRepositoryName, "primary")
}

func TestModifyManifests(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	mo := testutil.DummyManifestOptions()
	manifest := conf.AddManifest(mo)
	require.NotNil(t, manifest)

	mo.TargetPath += stringDelta
	err := conf.ModifyManifest(manifest, mo)
	require.NoError(t, err)

	mo.CommitHash = "11ded0"
	mo.Tag = "v1.0"
	err = conf.ModifyManifest(manifest, mo)
	require.Error(t, err, "Checkout mutually exclusive, use either: commit-hash, branch or tag")

	// error scenario
	mo.RepoName = "invalid"
	mo.URL = ""
	err = conf.ModifyManifest(manifest, mo)
	require.Error(t, err)
}

func TestGetDefaultEncryptionConfigs(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	encryptionConfigs := conf.GetEncryptionConfigs()
	require.NotNil(t, encryptionConfigs)
	// by default, we dont expect any encryption configs
	assert.Equal(t, 0, len(encryptionConfigs))
}

func TestModifyEncryptionConfigs(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	eco := testutil.DummyEncryptionConfigOptions()
	encryptionConfig := conf.AddEncryptionConfig(eco)
	require.NotNil(t, encryptionConfig)

	eco.KeySecretName += stringDelta
	conf.ModifyEncryptionConfig(encryptionConfig, eco)
	modifiedConfig := conf.EncryptionConfigs[eco.Name]
	assert.Equal(t, eco.KeySecretName, modifiedConfig.KeySecretName)

	eco.KeySecretNamespace += stringDelta
	conf.ModifyEncryptionConfig(encryptionConfig, eco)
	assert.Equal(t, eco.KeySecretNamespace, modifiedConfig.KeySecretNamespace)

	eco.EncryptionKeyPath += stringDelta
	conf.ModifyEncryptionConfig(encryptionConfig, eco)
	assert.Equal(t, eco.EncryptionKeyPath, modifiedConfig.EncryptionKeyPath)

	eco.DecryptionKeyPath += stringDelta
	conf.ModifyEncryptionConfig(encryptionConfig, eco)
	assert.Equal(t, eco.DecryptionKeyPath, modifiedConfig.DecryptionKeyPath)
}

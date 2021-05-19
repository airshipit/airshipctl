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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"opendev.org/airship/airshipctl/pkg/util"

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

	assert.Len(t, conf.Contexts, 3)
}

func TestPersistConfig(t *testing.T) {
	testDir, err := ioutil.TempDir("", "airship-test")
	require.NoError(t, err)
	configPath := filepath.Join(testDir, "config")
	err = config.CreateConfig(configPath, true)
	require.NoError(t, err)
	assert.FileExists(t, configPath)
	err = os.RemoveAll(testDir)
	require.NoError(t, err)
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
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
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

	conf.ManagementConfiguration[defaultString] = testutil.DummyManagementConfiguration()

	managementConfig, err := conf.CurrentContextManagementConfig()
	require.Error(t, err)
	assert.Nil(t, managementConfig)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].ManagementConfiguration = defaultString

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

func TestGetContexts(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	contexts := conf.GetContexts()
	assert.Len(t, contexts, 3)
}

func TestGetContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	context, err := conf.GetContext("def_ephemeral")
	require.NoError(t, err)

	assert.EqualValues(t, context.Manifest, "dummy_manifest")

	// Test Wrong Cluster
	_, err = conf.GetContext("unknown")
	assert.Error(t, err)
}

func TestAddContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyContextOptions()
	context := conf.AddContext(co.Name, config.SetContextManifest(co.Manifest),
		config.SetContextManagementConfig(co.ManagementConfiguration))
	assert.EqualValues(t, conf.Contexts[co.Name], context)
}

func TestModifyContext(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyContextOptions()
	context := conf.AddContext(co.Name, config.SetContextManifest(co.Manifest))

	co.Manifest += stringDelta
	conf.ModifyContext(context, config.SetContextManifest(co.Manifest))
	assert.EqualValues(t, conf.Contexts[co.Name].Manifest, co.Manifest)
	assert.EqualValues(t, conf.Contexts[co.Name], context)
}

func TestGetCurrentContext(t *testing.T) {
	t.Run("getCurrentContext", func(t *testing.T) {
		conf, cleanup := testutil.InitConfig(t)
		defer cleanup(t)

		conf.CurrentContext = currentContextName
		conf.Contexts[currentContextName].Manifest = defaultString

		context, err := conf.GetCurrentContext()
		require.NoError(t, err)
		assert.Equal(t, conf.Contexts[currentContextName], context)
	})
}

func TestCurrentContextManifest(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.Manifests[defaultString] = testutil.DummyManifest()

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	manifest, err := conf.CurrentContextManifest()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString], manifest)
}

func TestCurrentTargetPath(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.Manifests[defaultString] = testutil.DummyManifest()

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	targetPath, err := conf.CurrentContextTargetPath()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString].TargetPath, targetPath)
}

func TestCurrentPhaseRepositoryDir(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.Manifests[defaultString] = testutil.DummyManifest()

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	phaseRepoDir, err := conf.CurrentContextPhaseRepositoryDir()
	require.NoError(t, err)
	assert.Equal(t, util.GitDirNameFromURL(
		conf.Manifests[defaultString].Repositories[conf.Manifests[defaultString].PhaseRepositoryName].URL()),
		phaseRepoDir)

	conf.Manifests[defaultString].PhaseRepositoryName = "nonexisting"
	phaseRepoDir, err = conf.CurrentContextPhaseRepositoryDir()
	require.Error(t, err)
	assert.Equal(t, config.ErrMissingRepositoryName{RepoType: "phase"}, err)
	assert.Equal(t, "", phaseRepoDir)
}

func TestCurrentInventoryRepositoryDir(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.Manifests[defaultString] = testutil.DummyManifest()

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	invRepoDir, err := conf.CurrentContextInventoryRepositoryName()
	require.NoError(t, err)
	assert.Equal(t, util.GitDirNameFromURL(
		conf.Manifests[defaultString].Repositories[conf.Manifests[defaultString].PhaseRepositoryName].URL()),
		invRepoDir)

	conf.Manifests[defaultString].InventoryRepositoryName = "nonexisting"
	invRepoDir, err = conf.CurrentContextInventoryRepositoryName()
	require.Error(t, err)
	assert.Equal(t, config.ErrMissingRepositoryName{RepoType: "inventory"}, err)
	assert.Equal(t, "", invRepoDir)

	invRepoName := "inv-repo"
	invRepoURL := "/my-repository"
	conf.Manifests[defaultString].Repositories[invRepoName] = &config.Repository{URLString: invRepoURL}
	conf.Manifests[defaultString].InventoryRepositoryName = invRepoName
	invRepoDir, err = conf.CurrentContextInventoryRepositoryName()
	require.NoError(t, err)
	assert.Equal(t, util.GitDirNameFromURL(
		conf.Manifests[defaultString].Repositories[conf.Manifests[defaultString].InventoryRepositoryName].URL()),
		invRepoDir)
}

func TestManagementConfigurationByName(t *testing.T) {
	conf, cleanupConfig := testutil.InitConfig(t)
	defer cleanupConfig(t)

	conf.ManagementConfiguration[defaultString] = testutil.DummyManagementConfiguration()

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

func TestGetManifest(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	_, err := conf.GetManifest("dummy_manifest")
	require.NoError(t, err)

	// Test Wrong Manifest
	_, err = conf.GetManifest("unknown")
	assert.Error(t, err)
}

func TestGetManifests(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	conf.Manifests["dummy_manifest"] = testutil.DummyManifest()

	manifests := conf.GetManifests()
	require.NotNil(t, manifests)

	assert.EqualValues(t, manifests[0].PhaseRepositoryName, "primary")
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

func TestWorkDir(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)
	wd, err := conf.WorkDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, wd)
}

func TestAddManagementConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	managementConfig := conf.AddManagementConfig("new_mgmt_context", config.SetManagementConfigUseProxy(false))
	assert.EqualValues(t, conf.ManagementConfiguration["new_mgmt_context"], managementConfig)
}

func TestModifyManagementConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	managementConfig := conf.AddManagementConfig("modified_mgmt_config")

	conf.ModifyManagementConfig(managementConfig, config.SetManagementConfigSystemActionRetries(60))
	assert.EqualValues(t, conf.ManagementConfiguration["modified_mgmt_config"].SystemActionRetries,
		managementConfig.SystemActionRetries)
	assert.EqualValues(t, conf.ManagementConfiguration["modified_mgmt_config"], managementConfig)
}

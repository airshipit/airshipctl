/*
Copyright 2014 The Kubernetes Authors.

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
	newToken           = "dummy_token_changed"
	newPassword        = "dummy_password_changed"
	newCertificate     = "dummy_certificate_changed"
	newKey             = "dummy_key_changed"
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
			name:     "cluster",
			stringer: testutil.DummyCluster(),
		},
		{
			name:     "authinfo",
			stringer: testutil.DummyAuthInfo(),
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
			name:     "bootstrapinfo",
			stringer: testutil.DummyBootstrapInfo(),
		},
		{
			name:     "managementconfiguration",
			stringer: testutil.DummyManagementConfiguration(),
		},
		{
			name: "builder",
			stringer: &config.Builder{
				UserDataFileName:       "user-data",
				NetworkConfigFileName:  "netconfig",
				OutputMetadataFileName: "output-metadata.yaml",
			},
		},
		{
			name: "container",
			stringer: &config.Container{
				Volume:           "/dummy:dummy",
				Image:            "dummy_image:dummy_tag",
				ContainerRuntime: "docker",
			},
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

	assert.Len(t, conf.Clusters, 6)
	require.Contains(t, conf.Clusters, "def")
	assert.Len(t, conf.Clusters["def"].ClusterTypes, 2)
	assert.Len(t, conf.Contexts, 3)
	assert.Len(t, conf.AuthInfos, 3)
}

func TestPersistConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	err := conf.PersistConfig()
	require.NoError(t, err)

	// Check that the files were created
	assert.FileExists(t, conf.LoadedConfigPath())
	assert.FileExists(t, conf.KubeConfigPath())
	// Check that the invalid name was changed to a valid one
	assert.Contains(t, conf.KubeConfig().Clusters, "invalidName_target")

	// Check that the missing cluster was added to the airshipconfig
	assert.Contains(t, conf.Clusters, "onlyinkubeconf")

	// Check that the "stragglers" were removed from the airshipconfig
	assert.NotContains(t, conf.Clusters, "straggler")
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
			name: "no clusters defined",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one cluster needs to be defined"},
		},
		{
			name: "no users defined",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{},
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one Authentication Information (User) needs to be defined"},
		},
		{
			name: "no contexts defined",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one Context needs to be defined"},
		},
		{
			name: "no manifests defined",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "At least one Manifest needs to be defined"},
		},
		{
			name: "current context not defined",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context is not defined"},
		},
		{
			name: "no context for current context",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{"DIFFERENT_CONTEXT": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Context"},
		},
		{
			name: "no manifest for current context",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*config.Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*config.Manifest{"DIFFERENT_MANIFEST": {}},
				CurrentContext: "testContext",
			},
			expectedErr: config.ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Manifest"},
		},
		{
			name: "complete config",
			config: config.Config{
				Clusters:       map[string]*config.ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*config.AuthInfo{"testAuthInfo": {}},
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

func TestCurrentContextBootstrapInfo(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"
	clusterType := "ephemeral"

	bootstrapInfo, err := conf.CurrentContextBootstrapInfo()
	require.Error(t, err)
	assert.Nil(t, bootstrapInfo)

	conf.CurrentContext = currentContextName
	conf.Clusters[clusterName].ClusterTypes[clusterType].Bootstrap = defaultString
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

	bootstrapInfo, err = conf.CurrentContextBootstrapInfo()
	require.NoError(t, err)
	assert.Equal(t, conf.BootstrapInfo[defaultString], bootstrapInfo)
}

func TestCurrentContextManagementConfig(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"
	clusterType := "ephemeral"

	managementConfig, err := conf.CurrentContextManagementConfig()
	require.Error(t, err)
	assert.Nil(t, managementConfig)

	conf.CurrentContext = currentContextName
	conf.Clusters[clusterName].ClusterTypes[clusterType].ManagementConfiguration = defaultString
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

	managementConfig, err = conf.CurrentContextManagementConfig()
	require.NoError(t, err)
	assert.Equal(t, conf.ManagementConfiguration[defaultString], managementConfig)
}

func TestPurge(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	// Store it
	err := conf.PersistConfig()
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

func TestModifyCluster(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	co := testutil.DummyClusterOptions()
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)

	co.Server += stringDelta
	co.InsecureSkipTLSVerify = true
	co.EmbedCAData = true
	mcluster, err := conf.ModifyCluster(cluster, co)
	require.NoError(t, err)
	assert.EqualValues(t, conf.Clusters[co.Name].ClusterTypes[co.ClusterType].KubeCluster().Server, co.Server)
	assert.EqualValues(t, conf.Clusters[co.Name].ClusterTypes[co.ClusterType], mcluster)

	// Error case
	co.CertificateAuthority = "unknown"
	_, err = conf.ModifyCluster(cluster, co)
	assert.Error(t, err)
}

func TestGetClusters(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusters := conf.GetClusters()
	assert.Len(t, clusters, 6)
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

	// Test Positives
	assert.EqualValues(t, context.NameInKubeconf, "def_ephemeral")
	assert.EqualValues(t, context.KubeContext().Cluster, "def_ephemeral")

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

	co.Namespace += stringDelta
	co.Cluster += stringDelta
	co.AuthInfo += stringDelta
	co.Manifest += stringDelta
	conf.ModifyContext(context, co)
	assert.EqualValues(t, conf.Contexts[co.Name].KubeContext().Namespace, co.Namespace)
	assert.EqualValues(t, conf.Contexts[co.Name].KubeContext().Cluster, co.Cluster)
	assert.EqualValues(t, conf.Contexts[co.Name].KubeContext().AuthInfo, co.AuthInfo)
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

func TestCurrentContextCluster(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"
	clusterType := "ephemeral"

	cluster, err := conf.CurrentContextCluster()
	require.Error(t, err)
	assert.Nil(t, cluster)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

	cluster, err = conf.CurrentContextCluster()
	require.NoError(t, err)
	assert.Equal(t, conf.Clusters[clusterName].ClusterTypes[clusterType], cluster)
}

func TestCurrentContextAuthInfo(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	authInfo, err := conf.CurrentContextAuthInfo()
	require.Error(t, err)
	assert.Nil(t, authInfo)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString

	authInfo, err = conf.CurrentContextAuthInfo()
	require.NoError(t, err)
	assert.Equal(t, conf.AuthInfos["k-admin"], authInfo)
}

func TestCurrentContextManifest(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"

	manifest, err := conf.CurrentContextManifest()
	require.Error(t, err)
	assert.Nil(t, manifest)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

	manifest, err = conf.CurrentContextManifest()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString], manifest)
}

func TestCurrentTargetPath(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"

	manifest, err := conf.CurrentContextManifest()
	require.Error(t, err)
	assert.Nil(t, manifest)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

	targetPath, err := conf.CurrentContextTargetPath()
	require.NoError(t, err)
	assert.Equal(t, conf.Manifests[defaultString].TargetPath, targetPath)
}

func TestCurrentContextEntryPoint(t *testing.T) {
	conf, cleanup := testutil.InitConfig(t)
	defer cleanup(t)

	clusterName := "def"

	entryPoint, err := conf.CurrentContextEntryPoint(defaultString)
	require.Error(t, err)
	assert.Equal(t, "", entryPoint)

	conf.CurrentContext = currentContextName
	conf.Contexts[currentContextName].Manifest = defaultString
	conf.Contexts[currentContextName].KubeContext().Cluster = clusterName

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
			metaPath:       "testdata/metadata.yaml",
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
				TargetPath:   ".",
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

func TestImport(t *testing.T) {
	conf, cleanupConfig := testutil.InitConfig(t)
	defer cleanupConfig(t)

	kubeDir, cleanupKubeConfig := testutil.TempDir(t, "airship-import-tests")
	defer cleanupKubeConfig(t)

	kubeConfigPath := filepath.Join(kubeDir, "config")
	//nolint: lll
	kubeConfigContent := `
apiVersion: v1
clusters:
- cluster:
    server: https://1.2.3.4:9000
  name: cluster_target
- cluster:
    server: https://1.2.3.4:9001
  name: dummycluster_ephemeral
- cluster:
    server: https://1.2.3.4:9002
  name: def_target
- cluster:
    server: https://1.2.3.4:9003
  name: noncomplex
contexts:
- context:
    cluster: cluster_target
    user: cluster-admin
  name: cluster-admin@cluster
- context:
    cluster: dummycluster_ephemeral
    user: kubernetes-admin
  name: dummy_cluster
- context:
    cluster: dummycluster_ephemeral
    user: kubernetes-admin
  name: def_target
- context:
    cluster: noncomplex
    user: kubernetes-admin
  name: noncomplex
current-context: dummy_cluster
kind: Config
preferences: {}
users:
- name: cluster-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
- name: kubernetes-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
- name: def-user
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
`
	err := ioutil.WriteFile(kubeConfigPath, []byte(kubeConfigContent), 0644)
	require.NoError(t, err)

	err = conf.ImportFromKubeConfig(kubeConfigPath)
	require.NoError(t, err)

	t.Run("importClusters", func(t *testing.T) {
		// Verify that only 3 clusters have been added (original 5 plus 3 new clusters)
		// This is important since the above kubeconfig actually has 4
		// clusters, but one was already defined in the airship config
		assert.Len(t, conf.Clusters, 6+3)

		// verify that the new clusters have been added to the config
		_, err := conf.GetCluster("cluster", config.Target)
		assert.NoError(t, err)
		_, err = conf.GetCluster("dummycluster", config.Ephemeral)
		assert.NoError(t, err)

		// verify that the "noncomplex" cluster was added as a target cluster
		_, err = conf.GetCluster("noncomplex", config.Target)
		assert.NoError(t, err)
	})

	t.Run("importContexts", func(t *testing.T) {
		// Verify that only 3 contexts have been added (original 3 plus 3 new contexts)
		// This is important since the above kubeconfig actually has 4
		// contexts, but one was already defined in the airship config
		assert.Len(t, conf.Contexts, 3+3)

		// verify that the new contexts have been added to the config
		_, err := conf.GetContext("cluster-admin@cluster")
		assert.NoError(t, err)
		_, err = conf.GetContext("dummy_cluster")
		assert.NoError(t, err)

		// verify that the "noncomplex" context refers to the proper target "noncomplex" cluster
		noncomplex, err := conf.GetContext("noncomplex")
		require.NoError(t, err)
		assert.Equal(t, "noncomplex_target", noncomplex.NameInKubeconf)
	})

	t.Run("importAuthInfos", func(t *testing.T) {
		// Verify that only 2 users have been added (original 3 plus 2 new users)
		// This is important since the above kubeconfig actually has 3
		// users, but one was already defined in the airship config
		assert.Len(t, conf.AuthInfos, 3+2)

		// verify that the new users have been added to the config
		_, err := conf.GetAuthInfo("cluster-admin")
		assert.NoError(t, err)
		_, err = conf.GetAuthInfo("kubernetes-admin")
		assert.NoError(t, err)
	})
}

func TestImportErrors(t *testing.T) {
	conf, cleanupConfig := testutil.InitConfig(t)
	defer cleanupConfig(t)

	t.Run("nonexistent kubeConfig", func(t *testing.T) {
		err := conf.ImportFromKubeConfig("./non/existent/file/path")
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("malformed kubeConfig", func(t *testing.T) {
		kubeDir, cleanupKubeConfig := testutil.TempDir(t, "airship-import-tests")
		defer cleanupKubeConfig(t)

		kubeConfigPath := filepath.Join(kubeDir, "config")
		//nolint: lll
		kubeConfigContent := "malformed content"

		err := ioutil.WriteFile(kubeConfigPath, []byte(kubeConfigContent), 0644)
		require.NoError(t, err)

		err = conf.ImportFromKubeConfig(kubeConfigPath)
		assert.Contains(t, err.Error(), "json parse error")
	})
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

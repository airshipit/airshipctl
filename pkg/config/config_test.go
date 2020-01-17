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

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"opendev.org/airship/airshipctl/testutil"
)

func TestString(t *testing.T) {
	fSys := testutil.SetupTestFs(t, "testdata")

	tests := []struct {
		name     string
		stringer fmt.Stringer
	}{
		{
			name:     "config",
			stringer: DummyConfig(),
		},
		{
			name:     "context",
			stringer: DummyContext(),
		},
		{
			name:     "cluster",
			stringer: DummyCluster(),
		},
		{
			name:     "authinfo",
			stringer: DummyAuthInfo(),
		},
		{
			name:     "manifest",
			stringer: DummyManifest(),
		},
		{
			name:     "modules",
			stringer: DummyModules(),
		},
		{
			name:     "repository",
			stringer: DummyRepository(),
		},
		{
			name:     "bootstrap",
			stringer: DummyBootstrap(),
		},
		{
			name:     "bootstrap",
			stringer: DummyBootstrap(),
		},
		{
			name: "builder",
			stringer: &Builder{
				UserDataFileName:       "user-data",
				NetworkConfigFileName:  "netconfig",
				OutputMetadataFileName: "output-metadata.yaml",
			},
		},
		{
			name: "container",
			stringer: &Container{
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

func TestPrettyString(t *testing.T) {
	fSys := testutil.SetupTestFs(t, "testdata")
	data, err := fSys.ReadFile("/prettycluster-string.yaml")
	require.NoError(t, err)

	cluster := DummyCluster()
	assert.EqualValues(t, cluster.PrettyString(), string(data))
}

func TestEqual(t *testing.T) {
	t.Run("config-equal", func(t *testing.T) {
		testConfig1 := NewConfig()
		testConfig2 := NewConfig()
		testConfig2.Kind = "Different"
		assert.True(t, testConfig1.Equal(testConfig1))
		assert.False(t, testConfig1.Equal(testConfig2))
		assert.False(t, testConfig1.Equal(nil))
	})

	t.Run("cluster-equal", func(t *testing.T) {
		testCluster1 := &Cluster{NameInKubeconf: "same"}
		testCluster2 := &Cluster{NameInKubeconf: "different"}
		assert.True(t, testCluster1.Equal(testCluster1))
		assert.False(t, testCluster1.Equal(testCluster2))
		assert.False(t, testCluster1.Equal(nil))
	})

	t.Run("context-equal", func(t *testing.T) {
		testContext1 := &Context{NameInKubeconf: "same"}
		testContext2 := &Context{NameInKubeconf: "different"}
		assert.True(t, testContext1.Equal(testContext1))
		assert.False(t, testContext1.Equal(testContext2))
		assert.False(t, testContext1.Equal(nil))
	})

	// TODO(howell): this needs to be fleshed out when the AuthInfo type is finished
	t.Run("authinfo-equal", func(t *testing.T) {
		testAuthInfo1 := &AuthInfo{}
		assert.True(t, testAuthInfo1.Equal(testAuthInfo1))
		assert.False(t, testAuthInfo1.Equal(nil))
	})

	t.Run("manifest-equal", func(t *testing.T) {
		testManifest1 := &Manifest{TargetPath: "same"}
		testManifest2 := &Manifest{TargetPath: "different"}
		assert.True(t, testManifest1.Equal(testManifest1))
		assert.False(t, testManifest1.Equal(testManifest2))
		assert.False(t, testManifest1.Equal(nil))
	})

	t.Run("repository-equal", func(t *testing.T) {
		testRepository1 := &Repository{TargetPath: "same"}
		testRepository2 := &Repository{TargetPath: "different"}
		assert.True(t, testRepository1.Equal(testRepository1))
		assert.False(t, testRepository1.Equal(testRepository2))
		assert.False(t, testRepository1.Equal(nil))
	})

	t.Run("modules-equal", func(t *testing.T) {
		testModules1 := NewModules()
		testModules2 := NewModules()
		testModules2.BootstrapInfo["different"] = &Bootstrap{
			Container: &Container{Volume: "different"},
		}
		assert.True(t, testModules1.Equal(testModules1))
		assert.False(t, testModules1.Equal(testModules2))
		assert.False(t, testModules1.Equal(nil))
	})

	t.Run("bootstrap-equal", func(t *testing.T) {
		testBootstrap1 := &Bootstrap{
			Container: &Container{
				Image: "same",
			},
		}
		testBootstrap2 := &Bootstrap{
			Container: &Container{
				Image: "different",
			},
		}
		assert.True(t, testBootstrap1.Equal(testBootstrap1))
		assert.False(t, testBootstrap1.Equal(testBootstrap2))
		assert.False(t, testBootstrap1.Equal(nil))
	})

	t.Run("container-equal", func(t *testing.T) {
		testContainer1 := &Container{Image: "same"}
		testContainer2 := &Container{Image: "different"}
		assert.True(t, testContainer1.Equal(testContainer1))
		assert.False(t, testContainer1.Equal(testContainer2))
		assert.False(t, testContainer1.Equal(nil))
	})

	t.Run("builder-equal", func(t *testing.T) {
		testBuilder1 := &Builder{UserDataFileName: "same"}
		testBuilder2 := &Builder{UserDataFileName: "different"}
		assert.True(t, testBuilder1.Equal(testBuilder1))
		assert.False(t, testBuilder1.Equal(testBuilder2))
		assert.False(t, testBuilder1.Equal(nil))
	})
}

func TestLoadConfig(t *testing.T) {
	conf := InitConfig(t)
	require.NotEmpty(t, conf.String())

	assert.Len(t, conf.Clusters, 4)
	require.Contains(t, conf.Clusters, "def")
	assert.Len(t, conf.Clusters["def"].ClusterTypes, 2)
	assert.Len(t, conf.Contexts, 3)
	assert.Len(t, conf.AuthInfos, 2)
}

func TestPersistConfig(t *testing.T) {
	config := InitConfig(t)

	err := config.PersistConfig()
	assert.NoErrorf(t, err, "Unable to persist configuration expected at %v", config.LoadedConfigPath())

	kpo := config.LoadedPathOptions()
	assert.NotNil(t, kpo)
}

func TestEnsureComplete(t *testing.T) {
	// This test is intentionally verbose. Since a user of EnsureComplete
	// does not need to know about the order of validation, each test
	// object passed into EnsureComplete should have exactly one issue, and
	// be otherwise valid
	tests := []struct {
		name        string
		config      Config
		expectedErr error
	}{
		{
			name: "no clusters defined",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "At least one cluster needs to be defined"},
		},
		{
			name: "no users defined",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "At least one Authentication Information (User) needs to be defined"},
		},
		{
			name: "no contexts defined",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{},
				Manifests:      map[string]*Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "At least one Context needs to be defined"},
		},
		{
			name: "no manifests defined",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "At least one Manifest needs to be defined"},
		},
		{
			name: "current context not defined",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"testManifest": {}},
				CurrentContext: "",
			},
			expectedErr: ErrMissingConfig{What: "Current Context is not defined"},
		},
		{
			name: "no context for current context",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"DIFFERENT_CONTEXT": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"testManifest": {}},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Context"},
		},
		{
			name: "no manifest for current context",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"DIFFERENT_MANIFEST": {}},
				CurrentContext: "testContext",
			},
			expectedErr: ErrMissingConfig{What: "Current Context (testContext) does not identify a defined Manifest"},
		},
		{
			name: "complete config",
			config: Config{
				Clusters:       map[string]*ClusterPurpose{"testCluster": {}},
				AuthInfos:      map[string]*AuthInfo{"testAuthInfo": {}},
				Contexts:       map[string]*Context{"testContext": {Manifest: "testManifest"}},
				Manifests:      map[string]*Manifest{"testManifest": {}},
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

func TestPurge(t *testing.T) {
	config := InitConfig(t)

	// Point the config objects at a temporary directory
	tempDir, err := ioutil.TempDir("", "airship-test-purge")
	require.NoError(t, err)

	airConfigFile := filepath.Join(tempDir, AirshipConfig)
	kConfigFile := filepath.Join(tempDir, AirshipKubeConfig)
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	// Store it
	err = config.PersistConfig()
	assert.NoErrorf(t, err, "Unable to persist configuration expected at %v", config.LoadedConfigPath())

	// Verify that the file is there
	_, err = os.Stat(config.LoadedConfigPath())
	assert.Falsef(t, os.IsNotExist(err), "Test config was not persisted at %v, cannot validate Purge",
		config.LoadedConfigPath())

	// Delete it
	err = config.Purge()
	assert.NoErrorf(t, err, "Unable to Purge file at %v", config.LoadedConfigPath())

	// Verify its gone
	_, err = os.Stat(config.LoadedConfigPath())
	assert.Falsef(t, os.IsExist(err), "Purge failed to remove file at %v", config.LoadedConfigPath())
}

func TestKClusterString(t *testing.T) {
	conf := InitConfig(t)
	kClusters := conf.KubeConfig().Clusters
	for kClust := range kClusters {
		assert.NotEmpty(t, KClusterString(kClusters[kClust]))
	}
	assert.EqualValues(t, KClusterString(nil), "null\n")
}
func TestKContextString(t *testing.T) {
	conf := InitConfig(t)
	kContexts := conf.KubeConfig().Contexts
	for kCtx := range kContexts {
		assert.NotEmpty(t, KContextString(kContexts[kCtx]))
	}
	assert.EqualValues(t, KClusterString(nil), "null\n")
}
func TestComplexName(t *testing.T) {
	cName := "aCluster"
	ctName := Ephemeral
	clusterName := NewClusterComplexName()
	clusterName.WithType(cName, ctName)
	assert.EqualValues(t, cName+"_"+ctName, clusterName.Name())

	assert.EqualValues(t, cName, clusterName.ClusterName())
	assert.EqualValues(t, ctName, clusterName.ClusterType())

	cName = "bCluster"
	clusterName.SetClusterName(cName)
	clusterName.SetDefaultType()
	ctName = clusterName.ClusterType()
	assert.EqualValues(t, cName+"_"+ctName, clusterName.Name())
	assert.EqualValues(t, "clusterName:"+cName+", clusterType:"+ctName, clusterName.String())
}

func TestValidClusterTypeFail(t *testing.T) {
	err := ValidClusterType("Fake")
	assert.Error(t, err)
}
func TestGetCluster(t *testing.T) {
	conf := InitConfig(t)
	cluster, err := conf.GetCluster("def", Ephemeral)
	require.NoError(t, err)

	// Test Positives
	assert.EqualValues(t, cluster.NameInKubeconf, "def_ephemeral")
	assert.EqualValues(t, cluster.KubeCluster().Server, "http://5.6.7.8")

	// Test Wrong Cluster
	_, err = conf.GetCluster("unknown", Ephemeral)
	assert.Error(t, err)

	// Test Wrong Cluster Type
	_, err = conf.GetCluster("def", "Unknown")
	assert.Error(t, err)
}

func TestAddCluster(t *testing.T) {
	co := DummyClusterOptions()
	conf := InitConfig(t)
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)

	assert.EqualValues(t, conf.Clusters[co.Name].ClusterTypes[co.ClusterType], cluster)
}

func TestModifyluster(t *testing.T) {
	co := DummyClusterOptions()
	conf := InitConfig(t)
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)

	co.Server += "/changes"
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
	conf := InitConfig(t)
	clusters, err := conf.GetClusters()
	require.NoError(t, err)
	assert.Len(t, clusters, 4)
}

func TestReconcileClusters(t *testing.T) {
	testCluster := &kubeconfig.Cluster{
		Server:                "testServer",
		InsecureSkipTLSVerify: true,
	}

	testKubeConfig := &kubeconfig.Config{
		Clusters: map[string]*kubeconfig.Cluster{
			"invalidName":                        nil,
			"missingFromAirshipConfig_ephemeral": testCluster,
		},
	}

	testConfig := &Config{
		Clusters: map[string]*ClusterPurpose{
			"straggler": {
				map[string]*Cluster{
					"ephemeral": {
						NameInKubeconf: "notThere!",
						kCluster:       nil,
					},
				},
			},
		},
		kubeConfig: testKubeConfig,
	}
	updatedClusterNames, persistIt := testConfig.reconcileClusters()

	// Check that there are clusters that need to be updated in contexts
	expectedUpdatedClusterNames := map[string]string{"invalidName": "invalidName_target"}
	assert.Equal(t, expectedUpdatedClusterNames, updatedClusterNames)

	// Check that we need to update the config file
	assert.True(t, persistIt)

	// Check that the invalid name was changed to a valid one
	assert.NotContains(t, testKubeConfig.Clusters, "invalidName")
	assert.Contains(t, testKubeConfig.Clusters, "invalidName_target")

	// Check that the missing cluster was added to the airshipconfig
	missingCluster := testConfig.Clusters["missingFromAirshipConfig"]
	require.NotNil(t, missingCluster)
	require.NotNil(t, missingCluster.ClusterTypes)
	require.NotNil(t, missingCluster.ClusterTypes[Ephemeral])
	assert.Equal(t, "missingFromAirshipConfig_ephemeral",
		missingCluster.ClusterTypes[Ephemeral].NameInKubeconf)
	assert.Equal(t, testCluster, missingCluster.ClusterTypes[Ephemeral].kCluster)

	// Check that the "stragglers" were removed from the airshipconfig
	assert.NotContains(t, testConfig.Clusters, "straggler")
}

func TestGetContexts(t *testing.T) {
	conf := InitConfig(t)
	contexts, err := conf.GetContexts()
	require.NoError(t, err)
	assert.Len(t, contexts, 3)
}

func TestGetContext(t *testing.T) {
	conf := InitConfig(t)
	context, err := conf.GetContext("def_ephemeral")
	require.NoError(t, err)

	// Test Positives
	assert.EqualValues(t, context.NameInKubeconf, "def_ephemeral")
	assert.EqualValues(t, context.KubeContext().Cluster, "def_ephemeral")

	// Test Wrong Cluster
	_, err = conf.GetContext("unknown")
	assert.Error(t, err)
}

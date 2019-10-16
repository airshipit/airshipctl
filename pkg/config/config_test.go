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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"

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

	// TODO(howell): this needs to be fleshed out when the Modules type is finished
	t.Run("modules-equal", func(t *testing.T) {
		testModules1 := &Modules{Dummy: "same"}
		testModules2 := &Modules{Dummy: "different"}
		assert.True(t, testModules1.Equal(testModules1))
		assert.False(t, testModules1.Equal(testModules2))
		assert.False(t, testModules1.Equal(nil))
	})
}

func TestLoadConfig(t *testing.T) {
	// Shouuld have the defult in testdata
	// And copy it to the default prior to the test
	// Create from defaults using existing kubeconf
	conf := InitConfig(t)

	require.NotEmpty(t, conf.String())

	// Lets make sure that the contents is as expected
	// 4 Clusters
	// 2 Clusters Types
	// 3 Contexts
	// 2 Users
	assert.Len(t, conf.Clusters, 4)
	assert.Len(t, conf.Clusters["def"].ClusterTypes, 2)
	assert.Len(t, conf.Contexts, 3)
	assert.Len(t, conf.AuthInfos, 2)

}

func TestPersistConfig(t *testing.T) {
	config := InitConfig(t)
	defer Clean(config)
	airConfigFile := filepath.Join(testAirshipConfigDir, AirshipConfig)
	kConfigFile := filepath.Join(testAirshipConfigDir, AirshipKubeConfig)
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	err := config.PersistConfig()
	assert.NoErrorf(t, err, "Unable to persist configuration expected at %v", config.LoadedConfigPath())

	kpo := config.LoadedPathOptions()
	assert.NotNil(t, kpo)
}

func TestPersistConfigFail(t *testing.T) {
	config := InitConfig(t)
	defer Clean(config)

	airConfigFile := filepath.Join(testAirshipConfigDir, "\\")
	kConfigFile := filepath.Join(testAirshipConfigDir, "\\")
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	err := config.PersistConfig()
	assert.Errorf(t, err, "Able to persist configuration at %v, but expected an error", config.LoadedConfigPath())
}

func TestEnsureComplete(t *testing.T) {
	conf := InitConfig(t)

	conf.CurrentContext = "def_ephemeral"
	assert.NoError(t, conf.EnsureComplete())

	// Trigger no CurrentContext Error
	conf.CurrentContext = ""
	err := conf.EnsureComplete()
	assert.EqualError(t, err, "Config: Current Context is not defined, or it doesnt identify a defined Context")

	// Trigger Contexts Error
	for key := range conf.Contexts {
		delete(conf.Contexts, key)
	}
	err = conf.EnsureComplete()
	assert.EqualError(t, err, "Config: At least one Context needs to be defined")

	// Trigger Authentication Information
	for key := range conf.AuthInfos {
		delete(conf.AuthInfos, key)
	}
	err = conf.EnsureComplete()
	assert.EqualError(t, err, "Config: At least one Authentication Information (User) needs to be defined")

	conf = NewConfig()
	err = conf.EnsureComplete()
	assert.Error(t, err, "A new config object should not be complete")
}

func TestPurge(t *testing.T) {
	config := InitConfig(t)
	defer Clean(config)

	airConfigFile := filepath.Join(testAirshipConfigDir, AirshipConfig)
	kConfigFile := filepath.Join(testAirshipConfigDir, AirshipKubeConfig)
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	// Store it
	err := config.PersistConfig()
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

func TestClusterNames(t *testing.T) {
	conf := InitConfig(t)
	expected := []string{"def", "onlyinkubeconf", "wrongonlyinconfig", "wrongonlyinkubeconf"}
	assert.EqualValues(t, expected, conf.ClusterNames())
}
func TestKClusterString(t *testing.T) {
	conf := InitConfig(t)
	kClusters := conf.KubeConfig().Clusters
	for kClust := range kClusters {
		assert.NotEmpty(t, KClusterString(kClusters[kClust]))
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

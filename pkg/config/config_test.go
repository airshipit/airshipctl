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
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/client-go/tools/clientcmd"
)

// Testing related constants

var AirshipStructs = [...]reflect.Value{
	reflect.ValueOf(DummyConfig()),
	reflect.ValueOf(DummyCluster()),
	reflect.ValueOf(DummyContext()),
	reflect.ValueOf(DummyManifest()),
	reflect.ValueOf(DummyAuthInfo()),
	reflect.ValueOf(DummyRepository()),
	reflect.ValueOf(DummyModules()),
}

// I can probable reflect to generate this two slices, instead based on the 1st one
// Exercise left for later -- YES I will remove this comment in the next patchset
var AirshipStructsEqual = [...]reflect.Value{
	reflect.ValueOf(DummyConfig()),
	reflect.ValueOf(DummyCluster()),
	reflect.ValueOf(DummyContext()),
	reflect.ValueOf(DummyManifest()),
	reflect.ValueOf(DummyAuthInfo()),
	reflect.ValueOf(DummyRepository()),
	reflect.ValueOf(DummyModules()),
}

var AirshipStructsDiff = [...]reflect.Value{
	reflect.ValueOf(NewConfig()),
	reflect.ValueOf(NewCluster()),
	reflect.ValueOf(NewContext()),
	reflect.ValueOf(NewManifest()),
	reflect.ValueOf(NewAuthInfo()),
	reflect.ValueOf(NewRepository()),
	reflect.ValueOf(NewModules()),
}

// Test to complete min coverage
func TestString(t *testing.T) {
	for s := range AirshipStructs {
		airStruct := AirshipStructs[s]
		airStringMethod := airStruct.MethodByName("String")
		yaml := airStringMethod.Call([]reflect.Value{})
		require.NotNil(t, yaml)

		structName := strings.Split(airStruct.Type().String(), ".")
		expectedFile := filepath.Join(testDataDir, "GoldenString", structName[1]+testMimeType)
		expectedData, err := ioutil.ReadFile(expectedFile)
		assert.Nil(t, err)
		require.EqualValues(t, string(expectedData), yaml[0].String())

	}
}
func TestPrettyString(t *testing.T) {
	conf := InitConfig(t)
	cluster, err := conf.GetCluster("def", Ephemeral)
	require.NoError(t, err)
	expectedFile := filepath.Join(testDataDir, "GoldenString", "PrettyCluster.yaml")
	expectedData, err := ioutil.ReadFile(expectedFile)
	assert.Nil(t, err)

	assert.EqualValues(t, cluster.PrettyString(), string(expectedData))

}

func TestEqual(t *testing.T) {
	for s := range AirshipStructs {
		airStruct := AirshipStructs[s]
		airStringMethod := airStruct.MethodByName("Equal")
		args := []reflect.Value{AirshipStructsEqual[s]}
		eq := airStringMethod.Call(args)
		assert.NotNilf(t, eq, "Equal for %v failed to return response to Equal .  ", airStruct.Type().String())
		require.Truef(t, eq[0].Bool(), "Equal for %v failed to return true for equal values  ", airStruct.Type().String())

		// Lets test Equals against nil struct
		args = []reflect.Value{reflect.New(airStruct.Type()).Elem()}
		nileq := airStringMethod.Call(args)
		assert.NotNil(t, nileq, "Equal for %v failed to return response to Equal .  ", airStruct.Type().String())
		require.Falsef(t, nileq[0].Bool(),
			"Equal for %v failed to return false when comparing against nil value  ", airStruct.Type().String())

		// Ignore False Equals test for AuthInfo for now
		if airStruct.Type().String() == "*config.AuthInfo" {
			continue
		}
		// Lets test that equal returns false when they are diff
		args = []reflect.Value{AirshipStructsDiff[s]}
		neq := airStringMethod.Call(args)
		assert.NotNil(t, neq, "Equal for %v failed to return response to Equal .  ", airStruct.Type().String())
		require.Falsef(t, neq[0].Bool(),
			"Equal for %v failed to return false for different values  ", airStruct.Type().String())

	}
}

func TestLoadConfig(t *testing.T) {
	// Shouuld have the defult in testdata
	// And copy it to the default prior to the test
	// Create from defaults using existing kubeconf
	conf := InitConfig(t)

	require.NotEmpty(t, conf.String())

	// Lets make sure that the contents is as expected
	// 2 Clusters
	// 2 Clusters Types
	// 2 Contexts
	// 1 User
	require.Lenf(t, conf.Clusters, 4, "Expected 4 Clusters got %d", len(conf.Clusters))
	require.Lenf(t, conf.Clusters["def"].ClusterTypes, 2,
		"Expected 2 ClusterTypes got %d", len(conf.Clusters["def"].ClusterTypes))
	require.Len(t, conf.Contexts, 3, "Expected 3 Contexts got %d", len(conf.Contexts))
	require.Len(t, conf.AuthInfos, 2, "Expected 2 AuthInfo got %d", len(conf.AuthInfos))

}

func TestPersistConfig(t *testing.T) {
	config := InitConfig(t)
	airConfigFile := filepath.Join(testAirshipConfigDir, AirshipConfig)
	kConfigFile := filepath.Join(testAirshipConfigDir, AirshipKubeConfig)
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	err := config.PersistConfig()
	assert.Nilf(t, err, "Unable to persist configuration expected at  %v ", config.LoadedConfigPath())

	kpo := config.LoadedPathOptions()
	assert.NotNil(t, kpo)
	Clean(config)
}

func TestPersistConfigFail(t *testing.T) {
	config := InitConfig(t)
	airConfigFile := filepath.Join(testAirshipConfigDir, "\\")
	kConfigFile := filepath.Join(testAirshipConfigDir, "\\")
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	err := config.PersistConfig()
	require.NotNilf(t, err, "Able to persist configuration at %v expected an error", config.LoadedConfigPath())
	Clean(config)
}

func TestEnsureComplete(t *testing.T) {
	conf := InitConfig(t)

	err := conf.EnsureComplete()
	require.NotNilf(t, err, "Configuration was incomplete %v ", err.Error())

	// Trgger Contexts Error
	for key := range conf.Contexts {
		delete(conf.Contexts, key)
	}
	err = conf.EnsureComplete()
	assert.EqualValues(t, err.Error(), "Config: At least one Context needs to be defined")

	// Trigger Authentication Information
	for key := range conf.AuthInfos {
		delete(conf.AuthInfos, key)
	}
	err = conf.EnsureComplete()
	assert.EqualValues(t, err.Error(), "Config: At least one Authentication Information (User) needs to be defined")

	conf = NewConfig()
	err = conf.EnsureComplete()
	assert.NotNilf(t, err, "Configuration was found complete incorrectly")
}

func TestPurge(t *testing.T) {
	config := InitConfig(t)
	airConfigFile := filepath.Join(testAirshipConfigDir, AirshipConfig)
	kConfigFile := filepath.Join(testAirshipConfigDir, AirshipKubeConfig)
	config.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	config.SetLoadedPathOptions(kubePathOptions)

	// Store it
	err := config.PersistConfig()
	assert.Nilf(t, err, "Unable to persist configuration expected at  %v [%v] ",
		config.LoadedConfigPath(), err)

	// Verify that the file is there

	_, err = os.Stat(config.LoadedConfigPath())
	assert.Falsef(t, os.IsNotExist(err), "Test config was not persisted at  %v , cannot validate Purge [%v] ",
		config.LoadedConfigPath(), err)

	// Delete it
	err = config.Purge()
	assert.Nilf(t, err, "Unable to Purge file at  %v [%v] ", config.LoadedConfigPath(), err)

	// Verify its gone
	_, err = os.Stat(config.LoadedConfigPath())
	require.Falsef(t, os.IsExist(err), "Purge failed to remove file at  %v [%v] ",
		config.LoadedConfigPath(), err)

	Clean(config)
}

func TestClusterNames(t *testing.T) {
	conf := InitConfig(t)
	expected := []string{"def", "onlyinkubeconf", "wrongonlyinconfig", "wrongonlyinkubeconf"}
	require.EqualValues(t, expected, conf.ClusterNames())
}
func TestKClusterString(t *testing.T) {
	conf := InitConfig(t)
	kClusters := conf.KubeConfig().Clusters
	for kClust := range kClusters {
		require.NotNil(t, KClusterString(kClusters[kClust]))
	}
	require.EqualValues(t, KClusterString(nil), "null\n")
}
func TestComplexName(t *testing.T) {
	cName := "aCluster"
	ctName := Ephemeral
	clusterName := NewClusterComplexName()
	clusterName.WithType(cName, ctName)
	require.EqualValues(t, cName+"_"+ctName, clusterName.Name())

	require.EqualValues(t, cName, clusterName.ClusterName())
	require.EqualValues(t, ctName, clusterName.ClusterType())

	cName = "bCluster"
	clusterName.SetClusterName(cName)
	clusterName.SetDefaultType()
	ctName = clusterName.ClusterType()
	require.EqualValues(t, cName+"_"+ctName, clusterName.Name())

	require.EqualValues(t, "clusterName:"+cName+", clusterType:"+ctName, clusterName.String())
}

func TestValidClusterTypeFail(t *testing.T) {
	err := ValidClusterType("Fake")
	require.NotNil(t, err)
}
func TestGetCluster(t *testing.T) {
	conf := InitConfig(t)
	cluster, err := conf.GetCluster("def", Ephemeral)
	require.NoError(t, err)

	// Test Positives
	assert.EqualValues(t, cluster.NameInKubeconf, "def_ephemeral")
	assert.EqualValues(t, cluster.KubeCluster().Server, "http://5.6.7.8")
	// Test Wrong Cluster
	cluster, err = conf.GetCluster("unknown", Ephemeral)
	assert.NotNil(t, err)
	assert.Nil(t, cluster)
	// Test Wrong Cluster Type
	cluster, err = conf.GetCluster("def", "Unknown")
	assert.NotNil(t, err)
	assert.Nil(t, cluster)
	// Test Wrong Cluster Type

}
func TestAddCluster(t *testing.T) {
	co := DummyClusterOptions()
	conf := InitConfig(t)
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)
	require.NotNil(t, cluster)
	assert.EqualValues(t, conf.Clusters[co.Name].ClusterTypes[co.ClusterType], cluster)
}
func TestModifyluster(t *testing.T) {
	co := DummyClusterOptions()
	conf := InitConfig(t)
	cluster, err := conf.AddCluster(co)
	require.NoError(t, err)
	require.NotNil(t, cluster)

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
	assert.NotNil(t, err)
}

func TestGetClusters(t *testing.T) {
	conf := InitConfig(t)
	clusters, err := conf.GetClusters()
	require.NoError(t, err)
	assert.EqualValues(t, 4, len(clusters))

}

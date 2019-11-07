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
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/require"
)

const (
	testDataDir          = "../../pkg/config/testdata"
	testAirshipConfig    = "testconfig"
	testAirshipConfigDir = ".testairship"
	testMimeType         = ".yaml"
)

// DummyConfig used by tests, to initialize min set of data
func DummyConfig() *Config {
	conf := &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigApiVersion,
		Clusters: map[string]*ClusterPurpose{
			"dummy_cluster": DummyClusterPurpose(),
		},
		AuthInfos: map[string]*AuthInfo{
			"dummy_user": DummyAuthInfo(),
		},
		Contexts: map[string]*Context{
			"dummy_context": DummyContext(),
		},
		Manifests: map[string]*Manifest{
			"dummy_manifest": DummyManifest(),
		},
		ModulesConfig:  DummyModules(),
		CurrentContext: "dummy_context",
		kubeConfig:     kubeconfig.NewConfig(),
	}
	dummyCluster := conf.Clusters["dummy_cluster"]
	conf.KubeConfig().Clusters["dummycluster_target"] = dummyCluster.ClusterTypes[Target].KubeCluster()
	conf.KubeConfig().Clusters["dummycluster_ephemeral"] = dummyCluster.ClusterTypes[Ephemeral].KubeCluster()
	return conf
}

// DummyContext , utility function used for tests
func DummyContext() *Context {
	c := NewContext()
	c.NameInKubeconf = "dummy_cluster"
	c.Manifest = "dummy_manifest"
	return c
}

// DummyCluster, utility function used for tests
func DummyCluster() *Cluster {
	c := NewCluster()

	cluster := kubeconfig.NewCluster()
	cluster.Server = "http://dummy.server"
	cluster.InsecureSkipTLSVerify = false
	cluster.CertificateAuthority = "dummy_ca"
	c.SetKubeCluster(cluster)
	c.NameInKubeconf = "dummycluster_target"
	c.Bootstrap = "dummy_bootstrap"
	return c
}

// DummyManifest , utility function used for tests
func DummyManifest() *Manifest {
	m := NewManifest()
	// Repositories is the map of repository adddressable by a name
	m.Repositories["dummy"] = DummyRepository()
	m.TargetPath = "/var/tmp/"
	return m
}

func DummyRepository() *Repository {
	url, _ := url.Parse("http://dummy.url.com")
	return &Repository{
		Url:        url,
		Username:   "dummy_user",
		TargetPath: "dummy_targetpath",
	}
}

func DummyAuthInfo() *AuthInfo {
	return NewAuthInfo()
}

func DummyModules() *Modules {
	return &Modules{Dummy: "dummy-module"}
}

// DummyClusterPurpose , utility function used for tests
func DummyClusterPurpose() *ClusterPurpose {
	cp := NewClusterPurpose()
	cp.ClusterTypes["ephemeral"] = DummyCluster()
	cp.ClusterTypes["ephemeral"].NameInKubeconf = "dummycluster_ephemeral"
	cp.ClusterTypes["target"] = DummyCluster()
	return cp
}

func InitConfigAt(t *testing.T, airConfigFile, kConfigFile string) *Config {
	conf := NewConfig()

	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile

	err := conf.LoadConfig(airConfigFile, kubePathOptions)
	require.NoError(t, err)

	return conf
}
func InitConfig(t *testing.T) *Config {
	airConfigFile := filepath.Join(testDataDir, AirshipConfig+testMimeType)
	kConfigFile := filepath.Join(testDataDir, AirshipKubeConfig+testMimeType)
	return InitConfigAt(t, airConfigFile, kConfigFile)
}
func DefaultInitConfig(t *testing.T) *Config {
	conf := InitConfig(t)
	airConfigFile := filepath.Join(AirshipConfigDir, AirshipConfig)
	kConfigFile := filepath.Join(AirshipConfigDir, AirshipKubeConfig)
	conf.SetLoadedConfigPath(airConfigFile)
	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kConfigFile
	conf.SetLoadedPathOptions(kubePathOptions)
	return conf
}

func Clean(conf *Config) error {
	configDir := filepath.Dir(conf.LoadedConfigPath())
	err := os.RemoveAll(configDir)
	if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func DummyClusterOptions() *ClusterOptions {
	co := &ClusterOptions{}
	co.Name = "dummy_Cluster"
	co.ClusterType = Ephemeral
	co.Server = "http://1.1.1.1"
	co.InsecureSkipTLSVerify = false
	co.CertificateAuthority = ""
	co.EmbedCAData = false

	return co
}

/*
Copyright 2017 The Kubernetes Authors.

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
	"bytes"
	"io/ioutil"
	"testing"

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

type setClusterTest struct {
	description    string
	config         *config.Config
	args           []string
	flags          []string
	expected       string
	expectedConfig *config.Config
}

const (
	testCluster = "my-new-cluster"
)

func TestSetClusterWithCAFile(t *testing.T) {
	conf := config.InitConfig(t)
	certFile := "../../pkg/config/testdata/ca.crt"

	tname := testCluster
	tctype := config.Ephemeral

	expconf := config.InitConfig(t)
	expconf.Clusters[tname] = config.NewClusterPurpose()
	expconf.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expconf.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	expkCluster.CertificateAuthority = certFile
	expkCluster.InsecureSkipTLSVerify = false
	expconf.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagEmbedCerts + "=false",
			"--" + config.FlagCAFile + "=" + certFile,
			"--" + config.FlagInsecure + "=false",
		},
		expected:       `Cluster "` + tname + `" of type "` + config.Ephemeral + `" created.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}
func TestSetClusterWithCAFileData(t *testing.T) {
	conf := config.InitConfig(t)
	certFile := "../../pkg/config/testdata/ca.crt"

	tname := testCluster
	tctype := config.Ephemeral

	expconf := config.InitConfig(t)
	expconf.Clusters[tname] = config.NewClusterPurpose()
	expconf.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expconf.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	readData, err := ioutil.ReadFile(certFile)
	require.NoError(t, err)

	expkCluster.CertificateAuthorityData = readData
	expkCluster.InsecureSkipTLSVerify = false
	expconf.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagEmbedCerts + "=true",
			"--" + config.FlagCAFile + "=" + certFile,
			"--" + config.FlagInsecure + "=false",
		},
		expected:       `Cluster "` + tname + `" of type "` + config.Ephemeral + `" created.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}

func TestSetCluster(t *testing.T) {

	conf := config.InitConfig(t)

	//	err := conf.Purge()
	//	assert.Nilf(t, err, "Unable to purge test configuration %v", err)

	tname := testCluster
	tctype := config.Ephemeral

	expconf := config.InitConfig(t)
	expconf.Clusters[tname] = config.NewClusterPurpose()
	expconf.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expconf.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	expkCluster.Server = "https://192.168.0.11"
	expkCluster.InsecureSkipTLSVerify = false
	expconf.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagAPIServer + "=https://192.168.0.11",
			"--" + config.FlagInsecure + "=false",
		},
		expected:       `Cluster "` + tname + `" of type "` + config.Ephemeral + `" created.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}

func TestModifyCluster(t *testing.T) {
	tname := testClusterName
	tctype := config.Ephemeral

	conf := config.InitConfig(t)
	conf.Clusters[tname] = config.NewClusterPurpose()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	conf.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	conf.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()
	kCluster := kubeconfig.NewCluster()
	kCluster.Server = "https://192.168.0.10"
	conf.KubeConfig().Clusters[clusterName.Name()] = kCluster
	conf.Clusters[tname].ClusterTypes[tctype].SetKubeCluster(kCluster)

	expconf := config.InitConfig(t)
	expconf.Clusters[tname] = config.NewClusterPurpose()
	expconf.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	expconf.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()
	expkCluster := kubeconfig.NewCluster()
	expkCluster.Server = "https://192.168.0.10"
	expconf.KubeConfig().Clusters[clusterName.Name()] = expkCluster
	expconf.Clusters[tname].ClusterTypes[tctype].SetKubeCluster(expkCluster)

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with an existing cluster",
		config:      conf,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagAPIServer + "=https://192.168.0.99",
		},
		expected:       `Cluster "` + tname + `" of type "` + tctype + `" modified.` + "\n",
		expectedConfig: expconf,
	}
	test.run(t)
}

func (test setClusterTest) run(t *testing.T) {

	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.config)

	buf := bytes.NewBuffer([]byte{})

	cmd := NewCmdConfigSetCluster(settings)
	cmd.SetOutput(buf)
	cmd.SetArgs(test.args)
	err := cmd.Flags().Parse(test.flags)
	require.NoErrorf(t, err, "unexpected error flags args to command: %v,  flags: %v", err, test.flags)

	// Execute the Command
	// Which should Persist the File
	err = cmd.Execute()
	require.NoErrorf(t, err, "unexpected error executing command: %v, args: %v, flags: %v", err, test.args, test.flags)

	// Load a New Config from the default Config File
	//afterSettings := &environment.AirshipCTLSettings{}
	// Loads the Config File that was updated
	//afterSettings.NewConfig()
	// afterRunConf := afterSettings.GetConfig()
	afterRunConf := settings.Config()
	// Get ClusterType
	tctypeFlag := cmd.Flag(config.FlagClusterType)
	require.NotNil(t, tctypeFlag)
	tctype := tctypeFlag.Value.String()

	// Find the Cluster Created or Modified
	afterRunCluster, err := afterRunConf.GetCluster(test.args[0], tctype)
	require.NoError(t, err)
	require.NotNil(t, afterRunCluster)

	afterKcluster := afterRunCluster.KubeCluster()
	require.NotNil(t, afterKcluster)

	testKcluster := test.config.KubeConfig().
		Clusters[test.config.Clusters[test.args[0]].ClusterTypes[tctype].NameInKubeconf]

	assert.EqualValues(t, afterKcluster.Server, testKcluster.Server)

	// Test that the Return Message looks correct
	if len(test.expected) != 0 {
		assert.EqualValues(t, test.expected, buf.String())
	}
}

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
	"fmt"
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
	givenConfig    *config.Config
	args           []string
	flags          []string
	expectedOutput string
	expectedConfig *config.Config
}

const (
	testCluster = "my_new-cluster"
)

func TestSetClusterWithCAFile(t *testing.T) {
	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	certFile := "../../pkg/config/testdata/ca.crt"

	tname := testCluster
	tctype := config.Ephemeral

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Clusters[tname] = config.NewClusterPurpose()
	expected.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expected.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	expkCluster.CertificateAuthority = certFile
	expkCluster.InsecureSkipTLSVerify = false
	expected.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagEmbedCerts + "=false",
			"--" + config.FlagCAFile + "=" + certFile,
			"--" + config.FlagInsecure + "=false",
		},
		expectedOutput: fmt.Sprintf("Cluster %q of type %q created.\n", testCluster, config.Ephemeral),
		expectedConfig: expected,
	}
	test.run(t)
}
func TestSetClusterWithCAFileData(t *testing.T) {
	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	certFile := "../../pkg/config/testdata/ca.crt"

	tname := testCluster
	tctype := config.Ephemeral

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Clusters[tname] = config.NewClusterPurpose()
	expected.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expected.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	readData, err := ioutil.ReadFile(certFile)
	require.NoError(t, err)

	expkCluster.CertificateAuthorityData = readData
	expkCluster.InsecureSkipTLSVerify = false
	expected.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagEmbedCerts + "=true",
			"--" + config.FlagCAFile + "=" + certFile,
			"--" + config.FlagInsecure + "=false",
		},
		expectedOutput: fmt.Sprintf("Cluster %q of type %q created.\n", tname, config.Ephemeral),
		expectedConfig: expected,
	}
	test.run(t)
}

func TestSetCluster(t *testing.T) {
	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	tname := testCluster
	tctype := config.Ephemeral

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Clusters[tname] = config.NewClusterPurpose()
	expected.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	expected.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()

	expkCluster := kubeconfig.NewCluster()
	expkCluster.Server = "https://192.168.0.11"
	expkCluster.InsecureSkipTLSVerify = false
	expected.KubeConfig().Clusters[clusterName.Name()] = expkCluster

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with a new cluster",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagAPIServer + "=https://192.168.0.11",
			"--" + config.FlagInsecure + "=false",
		},
		expectedOutput: fmt.Sprintf("Cluster %q of type %q created.\n", tname, config.Ephemeral),
		expectedConfig: expected,
	}
	test.run(t)
}

func TestModifyCluster(t *testing.T) {
	tname := testClusterName
	tctype := config.Ephemeral

	given, cleanupGiven := config.InitConfig(t)
	defer cleanupGiven(t)

	given.Clusters[tname] = config.NewClusterPurpose()
	clusterName := config.NewClusterComplexName()
	clusterName.WithType(tname, tctype)
	given.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	given.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()
	cluster := kubeconfig.NewCluster()
	cluster.Server = "https://192.168.0.10"
	given.KubeConfig().Clusters[clusterName.Name()] = cluster
	given.Clusters[tname].ClusterTypes[tctype].SetKubeCluster(cluster)

	expected, cleanupExpected := config.InitConfig(t)
	defer cleanupExpected(t)

	expected.Clusters[tname] = config.NewClusterPurpose()
	expected.Clusters[tname].ClusterTypes[tctype] = config.NewCluster()
	expected.Clusters[tname].ClusterTypes[tctype].NameInKubeconf = clusterName.Name()
	expkCluster := kubeconfig.NewCluster()
	expkCluster.Server = "https://192.168.0.10"
	expected.KubeConfig().Clusters[clusterName.Name()] = expkCluster
	expected.Clusters[tname].ClusterTypes[tctype].SetKubeCluster(expkCluster)

	test := setClusterTest{
		description: "Testing 'airshipctl config set-cluster' with an existing cluster",
		givenConfig: given,
		args:        []string{tname},
		flags: []string{
			"--" + config.FlagClusterType + "=" + config.Ephemeral,
			"--" + config.FlagAPIServer + "=https://192.168.0.99",
		},
		expectedOutput: fmt.Sprintf("Cluster %q of type %q modified.\n", tname, tctype),
		expectedConfig: expected,
	}
	test.run(t)
}

func (test setClusterTest) run(t *testing.T) {
	// Get the Environment
	settings := &environment.AirshipCTLSettings{}
	settings.SetConfig(test.givenConfig)

	buf := bytes.NewBuffer([]byte{})

	cmd := NewCmdConfigSetCluster(settings)
	cmd.SetOut(buf)
	cmd.SetArgs(test.args)
	err := cmd.Flags().Parse(test.flags)
	require.NoErrorf(t, err, "unexpected error flags args to command: %v,  flags: %v", err, test.flags)

	// Execute the Command
	// Which should Persist the File
	err = cmd.Execute()
	require.NoErrorf(t, err, "unexpected error executing command: %v, args: %v, flags: %v", err, test.args, test.flags)

	// Loads the Config File that was updated
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

	testKcluster := test.givenConfig.KubeConfig().
		Clusters[test.givenConfig.Clusters[test.args[0]].ClusterTypes[tctype].NameInKubeconf]

	assert.EqualValues(t, afterKcluster.Server, testKcluster.Server)

	// Test that the Return Message looks correct
	if len(test.expectedOutput) != 0 {
		assert.EqualValues(t, test.expectedOutput, buf.String())
	}
}

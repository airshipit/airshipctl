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
	"testing"

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	ephemeralFlag = "--" + config.FlagClusterType + "=" + config.Ephemeral
	targetFlag    = "--" + config.FlagClusterType + "=" + config.Target

	fooCluster     = "clusterFoo"
	barCluster     = "clusterBar"
	bazCluster     = "clusterBaz"
	missingCluster = "clusterMissing"
)

func TestGetClusterCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{
		Config: &config.Config{
			Clusters: map[string]*config.ClusterPurpose{
				fooCluster: {
					ClusterTypes: map[string]*config.Cluster{
						config.Ephemeral: getNamedTestCluster(fooCluster, config.Ephemeral),
						config.Target:    getNamedTestCluster(fooCluster, config.Target),
					},
				},
				barCluster: {
					ClusterTypes: map[string]*config.Cluster{
						config.Ephemeral: getNamedTestCluster(barCluster, config.Ephemeral),
						config.Target:    getNamedTestCluster(barCluster, config.Target),
					},
				},
				bazCluster: {
					ClusterTypes: map[string]*config.Cluster{
						config.Ephemeral: getNamedTestCluster(bazCluster, config.Ephemeral),
						config.Target:    getNamedTestCluster(bazCluster, config.Target),
					},
				},
			},
		},
	}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-ephemeral",
			CmdLine: fmt.Sprintf("%s %s", ephemeralFlag, fooCluster),
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},
		{
			Name:    "get-target",
			CmdLine: fmt.Sprintf("%s %s", targetFlag, fooCluster),
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},

		// FIXME(howell): "airshipctl config get-cluster foo bar" will
		// print *all* clusters, regardless of whether they are
		// specified on the command line
		// In this case, the bazCluster should not be included in the
		// output, yet it is
		{
			Name:    "get-multiple-ephemeral",
			CmdLine: fmt.Sprintf("%s %s %s", ephemeralFlag, fooCluster, barCluster),
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},
		{
			Name:    "get-multiple-target",
			CmdLine: fmt.Sprintf("%s %s %s", targetFlag, fooCluster, barCluster),
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},

		// FIXME(howell): "airshipctl config get-cluster
		// --cluster-type=ephemeral" will print *all* clusters,
		// regardless of whether they are ephemeral or target
		{
			Name:    "get-all-ephemeral",
			CmdLine: ephemeralFlag,
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},
		{
			Name:    "get-all-target",
			CmdLine: targetFlag,
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
		},
		{
			Name:    "missing",
			CmdLine: fmt.Sprintf("%s %s", targetFlag, missingCluster),
			Cmd:     cmd.NewCmdConfigGetCluster(settings),
			Error: fmt.Errorf("cluster clustermissing information was not " +
				"found in the configuration"),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestNoClustersGetClusterCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{Config: new(config.Config)}
	cmdTest := &testutil.CmdTest{
		Name:    "no-clusters",
		CmdLine: "",
		Cmd:     cmd.NewCmdConfigGetCluster(settings),
	}
	testutil.RunTest(t, cmdTest)
}

func getNamedTestCluster(clusterName, clusterType string) *config.Cluster {
	kCluster := &kubeconfig.Cluster{
		LocationOfOrigin:      "",
		InsecureSkipTLSVerify: true,
		Server:                "",
	}

	newCluster := &config.Cluster{NameInKubeconf: fmt.Sprintf("%s_%s", clusterName, clusterType)}
	newCluster.SetKubeCluster(kCluster)

	return newCluster
}

package cluster_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd/cluster"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestNewClusterCommandReturn(t *testing.T) {
	fakeRootSettings := &environment.AirshipCTLSettings{
		AirshipConfigPath: "../../testdata/k8s/config.yaml",
		KubeConfigPath:    "../../testdata/k8s/kubeconfig.yaml",
	}
	fakeRootSettings.InitConfig()

	tests := []*testutil.CmdTest{
		{
			Name:    "cluster-cmd-with-defaults",
			CmdLine: "",
			Cmd:     cluster.NewClusterCommand(fakeRootSettings),
		},
		{
			Name:    "cluster-initinfra-cmd-with-defaults",
			CmdLine: "--help",
			Cmd:     cluster.NewCmdInitInfra(fakeRootSettings),
		},
	}
	for _, testcase := range tests {
		testutil.RunTest(t, testcase)
	}
}

package cluster_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd/cluster"
	"opendev.org/airship/airshipctl/testutil"
)

func TestNewClusterCommandReturn(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "cluster-cmd-with-defaults",
			CmdLine: "",
			Cmd:     cluster.NewClusterCommand(nil),
		},
	}
	for _, testcase := range tests {
		testutil.RunTest(t, testcase)
	}
}

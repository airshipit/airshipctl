package bootstrap

import (
	"testing"

	"opendev.org/airship/airshipctl/testutil"
)

func TestRemoteDirect(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "remotedirect-cmd-with-help",
			CmdLine: "remotedirect --help",
			Cmd:     NewRemoteDirectCommand(nil),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

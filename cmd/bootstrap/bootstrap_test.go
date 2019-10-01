package bootstrap_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd/bootstrap"
	"opendev.org/airship/airshipctl/testutil"
)

func TestBootstrap(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "bootstrap-cmd-with-defaults",
			CmdLine: "",
			Cmd:     bootstrap.NewBootstrapCommand(nil),
		},
		{
			Name:    "bootstrap-isogen-cmd-with-help",
			CmdLine: "isogen --help",
			Cmd:     bootstrap.NewBootstrapCommand(nil),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

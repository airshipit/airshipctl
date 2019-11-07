package cmd_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/testutil"
)

func TestVersion(t *testing.T) {
	versionCmd := cmd.NewVersionCommand()
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "version",
			CmdLine: "",
			Cmd:     versionCmd,
		},
		{
			Name:    "version-help",
			CmdLine: "--help",
			Cmd:     versionCmd,
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

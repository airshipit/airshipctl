package cmd_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/testutil"
)

func TestVersion(t *testing.T) {
	rootCmd, _, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	rootCmd.AddCommand(cmd.NewVersionCommand())

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "version",
			CmdLine: "version",
			Cmd:     rootCmd,
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

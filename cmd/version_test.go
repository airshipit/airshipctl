package cmd_test

import (
	"testing"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/test"
)

func TestVersion(t *testing.T) {
	cmdTests := []*test.CmdTest{
		{
			Name:    "version",
			CmdLine: "version",
		},
	}
	rootCmd, _, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	rootCmd.AddCommand(cmd.NewVersionCommand())
	for _, tt := range cmdTests {
		test.RunTest(t, tt, rootCmd)
	}
}

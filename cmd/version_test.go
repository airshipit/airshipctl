package cmd_test

import (
	"bytes"
	"testing"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/test"
)

func TestVersion(t *testing.T) {
	tt := test.CmdTest{
		Name:    "version",
		CmdLine: "version",
	}
	actual := &bytes.Buffer{}
	rootCmd, _, err := cmd.NewRootCmd(actual)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	rootCmd.AddCommand(cmd.NewVersionCommand(actual))
	test.RunTest(t, tt, rootCmd, actual)
}

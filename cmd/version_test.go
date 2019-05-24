package cmd_test

import (
	"bytes"
	"testing"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/internal/test"
)

func TestVersion(t *testing.T) {
	tests := []test.CmdTest{
		{
			Name:    "version",
			CmdLine: "version",
		},
	}
	for _, tt := range tests {
		actual := &bytes.Buffer{}
		rootCmd, err := cmd.NewRootCmd(actual)
		if err != nil {
			t.Fatalf("Could not create root command: %s", err.Error())
		}
		rootCmd.AddCommand(cmd.NewVersionCommand(actual))
		test.RunTest(t, tt, rootCmd, actual)
	}
}

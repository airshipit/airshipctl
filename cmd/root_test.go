package cmd_test

import (
	"testing"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/test"
)

func TestRoot(t *testing.T) {
	cmdTests := []*test.CmdTest{
		&test.CmdTest{
			Name:    "default",
			CmdLine: "",
		},
	}
	rootCmd, _, err := cmd.NewRootCmd(nil)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	for _, tt := range cmdTests {
		test.RunTest(t, tt, rootCmd)
	}
}

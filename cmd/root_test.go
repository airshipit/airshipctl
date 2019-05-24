package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/test"
)

func TestRoot(t *testing.T) {
	tt := test.CmdTest{
		Name:    "default",
		CmdLine: "",
	}
	actual := &bytes.Buffer{}
	rootCmd, _, err := cmd.NewRootCmd(actual)
	if err != nil {
		t.Fatalf("Could not create root command: %s", err.Error())
	}
	rootCmd.PersistentFlags().Parse(os.Args[1:])
	test.RunTest(t, tt, rootCmd, actual)
}

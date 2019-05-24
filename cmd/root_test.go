package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/test"
)

func TestRoot(t *testing.T) {
	tests := []test.CmdTest{
		{
			Name:    "default",
			CmdLine: "",
		},
	}
	for _, tt := range tests {
		actual := &bytes.Buffer{}
		rootCmd, err := cmd.NewRootCmd(actual)
		if err != nil {
			t.Fatalf("Could not create root command: %s", err.Error())
		}
		settings := &environment.AirshipCTLSettings{}
		settings.InitFlags(rootCmd)
		rootCmd.PersistentFlags().Parse(os.Args[1:])

		settings.Init()
		test.RunTest(t, tt, rootCmd, actual)
	}
}

package cmd_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/cmd/bootstrap"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRoot(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "rootCmd-with-no-defaults",
			CmdLine: "",
			Cmd:     getVanillaRootCmd(t),
		},
		{
			Name:    "rootCmd-with-defaults",
			CmdLine: "",
			Cmd:     getDefaultRootCmd(t),
		},
		{
			Name:    "specialized-rootCmd-with-bootstrap",
			CmdLine: "",
			Cmd:     getSpecializedRootCmd(t),
		},
	}

	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

func getVanillaRootCmd(t *testing.T) *cobra.Command {
	t.Helper()
	rootCmd, _, err := cmd.NewRootCmd(nil)
	require.NoError(t, err, "Could not create root commands")
	return rootCmd
}

func getDefaultRootCmd(t *testing.T) *cobra.Command {
	t.Helper()
	rootCmd, _, err := cmd.NewAirshipCTLCommand(nil)
	require.NoError(t, err, "Could not create root commands")
	return rootCmd
}

func getSpecializedRootCmd(t *testing.T) *cobra.Command {
	rootCmd := getVanillaRootCmd(t)
	rootCmd.AddCommand(bootstrap.NewBootstrapCommand(&environment.AirshipCTLSettings{}))
	return rootCmd
}

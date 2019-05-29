package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/cmd/bootstrap"
	"github.com/ian-howell/airshipctl/pkg/environment"
)

// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer) (*cobra.Command, *environment.AirshipCTLSettings, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipctl",
		Short: "airshipctl is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)
	rootCmd.AddCommand(NewVersionCommand(out))

	settings := &environment.AirshipCTLSettings{}
	settings.InitFlags(rootCmd)

	return rootCmd, settings, nil
}

// AddDefaultAirshipCTLCommands is a convenience function for adding all of the
// default commands to airshipctl
func AddDefaultAirshipCTLCommands(cmd *cobra.Command, settings *environment.AirshipCTLSettings) *cobra.Command {
	cmd.AddCommand(workflow.NewWorkflowCommand(cmd.OutOrStdout(), settings))
	cmd.AddCommand(bootstrap.NewBootstrapCommand(cmd.OutOrStdout(), settings))
	return cmd
}

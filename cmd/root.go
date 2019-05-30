package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/cmd/bootstrap"
	"github.com/ian-howell/airshipctl/cmd/workflow"
	"github.com/ian-howell/airshipctl/pkg/environment"
)

// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer) (*cobra.Command, *environment.AirshipCTLSettings, error) {
	settings := &environment.AirshipCTLSettings{}
	rootCmd := &cobra.Command{
		Use:           "airshipctl",
		Short:         "airshipctl is a unified entrypoint to various airship components",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := settings.Init(); err != nil {
				return fmt.Errorf("error while initializing settings: %s", err)
			}
			return nil
		},
	}
	rootCmd.SetOutput(out)
	rootCmd.AddCommand(NewVersionCommand())

	settings.InitFlags(rootCmd)

	return rootCmd, settings, nil
}

// AddDefaultAirshipCTLCommands is a convenience function for adding all of the
// default commands to airshipctl
func AddDefaultAirshipCTLCommands(cmd *cobra.Command, settings *environment.AirshipCTLSettings) *cobra.Command {
	cmd.AddCommand(workflow.NewWorkflowCommand(settings))
	cmd.AddCommand(bootstrap.NewBootstrapCommand(settings))
	return cmd
}

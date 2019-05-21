package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/log"
)


// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipctl",
		Short: "airshipctl is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)

	// Settings flags - This section should probably be moved to pkg/environment
	settings := &environment.AirshipCTLSettings{}
	settings.InitFlags(rootCmd)

	rootCmd.AddCommand(NewVersionCommand(out))

	loadPluginCommands(rootCmd, out, settings, args)

	rootCmd.PersistentFlags().Parse(args)

	log.Init(settings, out)


	return rootCmd, nil
}

// Execute runs the base airshipctl command
func Execute(out io.Writer) {
	rootCmd, err := NewRootCmd(out, os.Args[1:])
	if err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
}

// loadPluginCommands loads all of the plugins as subcommands to cmd
func loadPluginCommands(cmd *cobra.Command, out io.Writer, settings *environment.AirshipCTLSettings, args []string) {
	for _, subcmd := range pluginCommands {
		cmd.AddCommand(subcmd(out, settings, args))
	}
}

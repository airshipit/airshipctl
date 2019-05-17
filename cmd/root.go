package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/log"

	"github.com/spf13/cobra"
)

var settings environment.AirshipCTLSettings

// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer, args []string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipctl",
		Short: "airshipctl is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)

	// Settings flags - This section should probably be moved to pkg/environment
	rootCmd.PersistentFlags().BoolVar(&settings.Debug, "debug", false, "enable verbose output")

	rootCmd.AddCommand(NewVersionCommand(out))

	loadPluginCommands(rootCmd, out, args)

	log.Init(&settings, out)

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

// loadPluginCommands finds all of the plugins in the builtinPlugins and
// externalPlugins datastructures, and loads them as subcommands to cmd
func loadPluginCommands(cmd *cobra.Command, out io.Writer, args []string) {
	for _, subcmd := range builtinPlugins {
		cmd.AddCommand(subcmd(out, args))
	}

	for _, subcmd := range externalPlugins {
		cmd.AddCommand(subcmd(out, args))
	}
}

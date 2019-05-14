package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/log"
	"github.com/ian-howell/airshipctl/pkg/plugin"
	"github.com/ian-howell/airshipctl/pkg/util"

	"github.com/spf13/cobra"
)

var settings environment.AirshipCTLSettings

const defaultPluginDir = "_plugins/bin"

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

	if _, err := os.Stat(defaultPluginDir); err == nil {
		pluginFiles, err := util.ReadDir(defaultPluginDir)
		if err != nil {
			return nil, errors.New("could not read plugins: " + err.Error())
		}
		for _, pluginFile := range pluginFiles {
			pathToPlugin := filepath.Join(defaultPluginDir, pluginFile.Name())
			rootCmd.AddCommand(plugin.CreateCommandFromPlugin(pathToPlugin, out, args))
		}
	}

	log.Init(&settings, out)

	return rootCmd, nil
}

// Execute runs the base airshipctl command
func Execute(out io.Writer) {
	rootCmd, err := NewRootCmd(out, os.Args[1:])
	osExitIfError(out, err)
	osExitIfError(out, rootCmd.Execute())
}

func osExitIfError(out io.Writer, err error) {
	if err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
}

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/kube"
	"github.com/ian-howell/airshipctl/pkg/log"
	"github.com/ian-howell/airshipctl/pkg/plugin"

	"github.com/spf13/cobra"
)

var settings environment.AirshipCTLSettings

// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer, client *kube.Client, args []string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipctl",
		Short: "airshipctl is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)

	// Settings flags - This section should probably be moved to pkg/environment
	rootCmd.PersistentFlags().StringVar(&settings.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	rootCmd.PersistentFlags().BoolVar(&settings.Debug, "debug", false, "enable verbose output")
	if err := rootCmd.PersistentFlags().Parse(args); err != nil {
		return nil, errors.New("could not parse flags: " + err.Error())
	}

	log.Init(&settings, out)

	rootCmd.AddCommand(NewVersionCommand(out, client))

	workflowPlugin := "plugins/internal/workflow.so"
	if _, err := os.Stat(workflowPlugin); err == nil {
		rootCmd.AddCommand(plugin.CreateCommandFromPlugin(workflowPlugin, out, settings.KubeConfigFilePath))
	}

	return rootCmd, nil
}

// Execute runs the base airshipctl command
func Execute(out io.Writer) {
	// TODO(howell): Fix this unused error
	client, _ := kube.NewForConfig(settings.KubeConfigFilePath)

	rootCmd, err := NewRootCmd(out, client, os.Args[1:])
	osExitIfError(out, err)
	osExitIfError(out, rootCmd.Execute())
}

func osExitIfError(out io.Writer, err error) {
	if err != nil {
		fmt.Fprintln(out, err)
		os.Exit(1)
	}
}

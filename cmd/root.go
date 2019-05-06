package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ian-howell/airshipadm/pkg/environment"
	"github.com/ian-howell/airshipadm/pkg/kube"
	"github.com/ian-howell/airshipadm/pkg/log"
	"github.com/spf13/cobra"
)

var settings environment.AirshipADMSettings

// NewRootCmd creates the root `airshipadm` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer, client *kube.Client, args []string) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipadm",
		Short: "airshipadm is a unified entrypoint to various airship components",
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

	// Compound commands
	rootCmd.AddCommand(NewWorkflowCommand())
	return rootCmd, nil
}

// Execute runs the base airshipadm command
func Execute(out io.Writer) {
	client, err := kube.NewForConfig(settings.KubeConfigFilePath)
	osExitIfError(out, err)
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

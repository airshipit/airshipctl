package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ian-howell/airshipadm/pkg/environment"
	"github.com/ian-howell/airshipadm/pkg/kube"
	"github.com/spf13/cobra"
)

var settings environment.AirshipADMSettings

// NewRootCmd creates the root `airshipadm` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer, client *kube.Client, args []string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "airshipadm",
		Short: "airshipadm is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)

	// Settings flags - This section should probably be moved to pkg/environment
	rootCmd.PersistentFlags().StringVar(&settings.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	// TODO(howell): Remove this panic
	if err := rootCmd.PersistentFlags().Parse(args); err != nil {
		panic(err.Error())
	}

	rootCmd.AddCommand(NewVersionCommand(out, client))

	// Compound commands
	rootCmd.AddCommand(NewWorkflowCommand())

	return rootCmd
}

// Execute runs the base airshipadm command
func Execute(out io.Writer) {
	// TODO(howell): Remove this panic
	client, err := kube.NewForConfig(settings.KubeConfigFilePath)
	if err != nil {
		panic(err.Error())
	}

	rootCmd := NewRootCmd(out, client, os.Args[1:])
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

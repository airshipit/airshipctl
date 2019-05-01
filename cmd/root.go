package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/ian-howell/airshipadm/pkg/settings"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root `airshipadm` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "airshipadm",
		Short: "airshipadm is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)

	rootCmd.AddCommand(NewVersionCommand(out))

	// Compound commands
	rootCmd.AddCommand(NewWorkflowCommand())

	rootCmd.PersistentFlags().StringVar(&settings.Settings.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")

	return rootCmd
}

// Execute runs the base airshipadm command
func Execute(out io.Writer) {
	rootCmd := NewRootCmd(out)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

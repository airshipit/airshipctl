package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "airshipadm",
		Short:        "airshipadm is a unified entrypoint to various airship components",
	}


	rootCmd.AddCommand(NewVersionCommand())

	// Compound commands
	rootCmd.AddCommand(NewWorkflowCommand())

	return rootCmd
}

// Execute runs the base airshipadm command
func Execute() {
	rootCmd := NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

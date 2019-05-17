package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

var kubeConfigFilePath string

func NewWorkflowCommand(out io.Writer, args []string) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "access to workflows",
		Aliases: []string{"workflows", "wf"},
	}

	workflowRootCmd.PersistentFlags().StringVar(&kubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	workflowRootCmd.AddCommand(NewWorkflowListCommand(out, args))

	return workflowRootCmd
}

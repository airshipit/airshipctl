package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

var (
	kubeConfigFilePath string
	namespace          string
)

// NewWorkflowCommand creates a new command for working with argo workflows
func NewWorkflowCommand(out io.Writer, args []string) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Access to argo workflows",
		Aliases: []string{"workflows", "wf"},
	}

	workflowRootCmd.PersistentFlags().StringVar(&kubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	workflowRootCmd.PersistentFlags().StringVar(&namespace, "namespace", "default", "kubernetes namespace to use for the context of this command")

	workflowRootCmd.AddCommand(NewWorkflowInitCommand(out, args))
	workflowRootCmd.AddCommand(NewWorkflowListCommand(out, args))

	return workflowRootCmd
}

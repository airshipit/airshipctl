package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

func NewWorkflowCommand(out io.Writer) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "access to workflows",
		Aliases: []string{"workflows", "wf"},
	}

	workflowRootCmd.AddCommand(NewWorkflowListCommand(out))

	return workflowRootCmd
}

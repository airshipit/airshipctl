package cmd

import (
	"github.com/spf13/cobra"
)

func NewWorkflowCommand() *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "access to workflows",
		Aliases: []string{"workflows", "wf"},
	}

	return workflowRootCmd
}

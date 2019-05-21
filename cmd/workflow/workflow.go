package workflow

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
)

// NewWorkflowCommand creates a new command for working with argo workflows
func NewWorkflowCommand(out io.Writer, settings *environment.AirshipCTLSettings, args []string) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Access to argo workflows",
		Aliases: []string{"workflows", "wf"},
	}

	workflowRootCmd.AddCommand(NewWorkflowInitCommand(out, settings, args))
	workflowRootCmd.AddCommand(NewWorkflowListCommand(out, settings, args))

	return workflowRootCmd
}

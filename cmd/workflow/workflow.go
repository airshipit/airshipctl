package workflow

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

// PluginSettingsID is used as a key in the root settings map of plugin settings
const PluginSettingsID = "argo"

// NewWorkflowCommand creates a new command for working with argo workflows
func NewWorkflowCommand(out io.Writer, rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Access to argo workflows",
		Aliases: []string{"workflows", "wf"},
	}

	wfSettings := &wfenv.Settings{}
	wfSettings.InitFlags(workflowRootCmd)
	workflowRootCmd.AddCommand(NewWorkflowInitCommand(out, wfSettings))
	workflowRootCmd.AddCommand(NewWorkflowListCommand(out, rootSettings))
	rootSettings.Register(PluginSettingsID, wfSettings)

	return workflowRootCmd
}

package workflow

import (
	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

// PluginSettingsID is used as a key in the root settings map of plugin settings
const PluginSettingsID = "argo"

// NewWorkflowCommand creates a new command for working with argo workflows
func NewWorkflowCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "workflow",
		Short:   "Access to argo workflows",
		Aliases: []string{"workflows", "wf"},
	}

	wfSettings := &wfenv.Settings{
		AirshipCTLSettings: rootSettings,
	}
	wfSettings.InitFlags(workflowRootCmd)

	workflowRootCmd.AddCommand(NewWorkflowInitCommand(wfSettings))
	workflowRootCmd.AddCommand(NewWorkflowListCommand(wfSettings))

	return workflowRootCmd
}

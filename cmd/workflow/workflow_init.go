package workflow

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
	"github.com/ian-howell/airshipctl/pkg/workflow"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

var (
	manifestPath string
)

// NewWorkflowInitCommand is a command for bootstrapping a kubernetes cluster with the necessary components for Argo workflows
func NewWorkflowInitCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {

	workflowInitCommand := &cobra.Command{
		Use:   "init [flags]",
		Short: "bootstraps the kubernetes cluster with the Workflow CRDs and controller",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			wfSettings, ok := rootSettings.PluginSettings[PluginSettingsID].(*wfenv.Settings)
			if !ok {
				fmt.Fprintf(out, "settings for %s were not registered\n", PluginSettingsID)
				return
			}

			if err := workflow.Initialize(out, wfSettings, manifestPath); err != nil {
				fmt.Fprintf(out, "error while initializing argo: %s\n", err.Error())
				return
			}
		},
	}

	workflowInitCommand.PersistentFlags().StringVar(&manifestPath, "manifest", "", "path to a YAML manifest containing definitions of objects needed for Argo workflows")
	return workflowInitCommand
}

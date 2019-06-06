package workflow

import (
	"fmt"

	"github.com/spf13/cobra"

	wf "github.com/ian-howell/airshipctl/pkg/workflow"
	"github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

var (
	manifestPath string
)

// NewWorkflowInitCommand is a command for bootstrapping a kubernetes cluster with the necessary components for Argo workflows
func NewWorkflowInitCommand(settings *environment.Settings) *cobra.Command {

	workflowInitCommand := &cobra.Command{
		Use:   "init [flags]",
		Short: "bootstraps the kubernetes cluster with the Workflow CRDs and controller",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()

			clientset, err := wf.GetClientset(settings)
			if err != nil {
				fmt.Fprintf(out, "Could not get Workflow Clientset: %s\n", err.Error())
				return
			}

			if err := wf.Initialize(clientset, settings, manifestPath); err != nil {
				fmt.Fprintf(out, "error while initializing argo: %s\n", err.Error())
				return
			}
		},
	}

	workflowInitCommand.PersistentFlags().StringVar(&manifestPath, "manifest", "", "path to a YAML manifest containing definitions of objects needed for Argo workflows")
	return workflowInitCommand
}

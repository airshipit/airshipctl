package workflow

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ian-howell/airshipctl/pkg/environment"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

// NewWorkflowListCommand is a command for listing argo workflows
func NewWorkflowListCommand(out io.Writer, rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	workflowListCmd := &cobra.Command{
		Use:     "list",
		Short:   "list workflows",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			wfSettings, ok := rootSettings.PluginSettings[PluginSettingsID].(*wfenv.Settings)
			if !ok {
				fmt.Fprintf(out, "settings for %s were not registered\n", PluginSettingsID)
				return
			}
			clientSet := wfSettings.ArgoClient
			wflist, err := clientSet.Workflows(wfSettings.Namespace).List(v1.ListOptions{})
			if err != nil {
				panic(err.Error())
			}
			w := tabwriter.NewWriter(out, 0, 0, 5, ' ', 0)
			defer w.Flush()
			fmt.Fprintf(w, "%s\t%s\n", "NAME", "PHASE")
			for _, wf := range wflist.Items {
				fmt.Fprintf(w, "%s\t%s\n", wf.Name, wf.Status.Phase)
			}
		},
	}

	return workflowListCmd
}

package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// NewWorkflowListCommand is a command for listing argo workflows
func NewWorkflowListCommand(out io.Writer, args []string) *cobra.Command {
	workflowRootCmd := &cobra.Command{
		Use:     "list",
		Short:   "list workflows",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			if kubeConfigFilePath == "" {
				kubeConfigFilePath = clientcmd.RecommendedHomeFile
			}
			config, err := clientcmd.BuildConfigFromFlags("", kubeConfigFilePath)
			if err != nil {
				panic(err.Error())
			}

			clientSet, err := v1alpha1.NewForConfig(config)
			if err != nil {
				panic(err.Error())
			}

			wflist, err := clientSet.Workflows(namespace).List(v1.ListOptions{})
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

	return workflowRootCmd
}

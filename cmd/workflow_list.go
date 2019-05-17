package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewWorkflowListCommand(out io.Writer, args []string) *cobra.Command {

	// TODO(howell): This is only used to appease the linter. It will be used later
	_ = args

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

			crdConfig := *config
			crdConfig.ContentConfig.GroupVersion = &v1alpha1.SchemeGroupVersion
			crdConfig.APIPath = "/apis"
			crdConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
			crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

			exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
			if err != nil {
				panic(err)
			}

			wflist := v1alpha1.WorkflowList{}
			err = exampleRestClient.
				Get().
				Resource("workflows").
				Do().
				Into(&wflist)
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

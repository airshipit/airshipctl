package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	kube "github.com/ian-howell/airshipadm/pkg/kube"
)

const versionLong = `Show the versions for the airshipadm tool and its components.
This includes the following components, in order:
  * airshipadm client
  * kubernetes cluster
`

// NewVersionCommand prints out the versions of airshipadm and its underlying tools
func NewVersionCommand(out io.Writer, client *kube.Client) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipadm and its underlying tools",
		Long:  versionLong,
		Run: func(cmd *cobra.Command, args []string) {
			w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
			fmt.Fprintf(w, "%s:\t%s\n", "client", clientVersion())
			fmt.Fprintf(w, "%s:\t%s\n", "kubernetes server", kubeVersion(client))
			w.Flush()
		},
	}
	return versionCmd
}

func clientVersion() string {
	// TODO(howell): There's gotta be a smarter way to do this
	return "v0.1.0"
}

func kubeVersion(client *kube.Client) string {
	version, err := client.Discovery().ServerVersion()
	// TODO(howell): Remove this panic
	if err != nil {
		panic(err.Error())
	}
	return version.String()
}

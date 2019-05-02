package cmd

import (
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"

	kube "github.com/ian-howell/airshipadm/pkg/kube"
)

const versionLong = `Show the versions for the airshipadm tool and the supporting tools.
This includes the following tools, in order:
  * airshipadm
  * golang
  * kubernetes
  * helm
  * argo
`

// NewVersionCommand prints out the versions of airshipadm and its underlying tools
func NewVersionCommand(out io.Writer, client *kube.Client) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipadm and its underlying tools",
		Long:  versionLong,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(out, "%-10s: %s\n", "airshipadm", airshipadmVersion())
			fmt.Fprintf(out, "%-10s: %s\n", "golang", runtime.Version())
			fmt.Fprintf(out, "%-10s: %s\n", "kubernetes", kubeVersion(client))
			fmt.Fprintf(out, "%-10s: %s\n", "helm", helmVersion())
			fmt.Fprintf(out, "%-10s: %s\n", "argo", argoVersion())
		},
	}
	return versionCmd
}

func airshipadmVersion() string {
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

func helmVersion() string {
	// TODO(howell): Implement this
	return "TODO"
}

func argoVersion() string {
	// TODO(howell): Implement this
	return "TODO"
}

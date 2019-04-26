package cmd

import (
	"fmt"
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
func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipadm and its underlying tools",
		Long:  versionLong,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%-10s: %s\n", "airshipadm", airshipadmVersion())
			fmt.Printf("%-10s: %s\n", "golang", runtime.Version())
			fmt.Printf("%-10s: %s\n", "kubernetes", kubeVersion())
			fmt.Printf("%-10s: %s\n", "helm", helmVersion())
			fmt.Printf("%-10s: %s\n", "argo", argoVersion())
		},
	}
	return versionCmd
}

func airshipadmVersion() string {
	// TODO(howell): There's gotta be a smarter way to do this
	return "v0.1.0"
}

func kubeVersion() string {
	clientset := kube.GetClient()
	v, err := clientset.Discovery().ServerVersion()
	if err != nil {
		panic(err.Error())
	}
	return v.String()
}

func helmVersion() string {
	// TODO(howell): Implement this
	return "TODO"
}

func argoVersion() string {
	// TODO(howell): Implement this
	return "TODO"
}

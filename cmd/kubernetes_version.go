package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	k8s "github.com/ian-howell/airshipadm/pkg/kubernetes"
)

func NewKubernetesVersionCommand() *cobra.Command {
	kubernetesVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of the cluster",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(getVersion())
		},
	}

	return kubernetesVersionCmd
}

func getVersion() string {
	clientset := k8s.GetClient()

	v, err := clientset.Discovery().ServerVersion()
	if err != nil {
		panic(err.Error())
	}
	return v.String()
}

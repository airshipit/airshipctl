package cmd

import (
	"github.com/spf13/cobra"
)

func NewKubernetesCommand() *cobra.Command {
	kubernetesRootCmd := &cobra.Command{
		Use:   "kubernetes",
		Short: "access to various kubernetes tools",
		Aliases: []string{"kubectl", "k8s"},
	}

	kubernetesRootCmd.AddCommand(NewKubernetesVersionCommand())

	return kubernetesRootCmd
}

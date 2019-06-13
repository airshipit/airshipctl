package argo

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/argoproj/argo/util/cmd"
)

const (
	// CLIName is the name of the CLI
	CLIName = "argo"
)

// NewArgoCommand returns a new instance of an argo command
func NewArgoCommand() *cobra.Command {
	var pluginRootCmd = &cobra.Command{
		Use:   CLIName,
		Short: "argo is the command line interface to Argo",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.HelpFunc()(cmd, args)
		},
	}

	pluginRootCmd.AddCommand(NewDeleteCommand())
	pluginRootCmd.AddCommand(NewGetCommand())
	pluginRootCmd.AddCommand(NewLintCommand())
	pluginRootCmd.AddCommand(NewListCommand())
	pluginRootCmd.AddCommand(NewLogsCommand())
	pluginRootCmd.AddCommand(NewResubmitCommand())
	pluginRootCmd.AddCommand(NewResumeCommand())
	pluginRootCmd.AddCommand(NewRetryCommand())
	pluginRootCmd.AddCommand(NewSubmitCommand())
	pluginRootCmd.AddCommand(NewSuspendCommand())
	pluginRootCmd.AddCommand(NewWaitCommand())
	pluginRootCmd.AddCommand(NewWatchCommand())
	pluginRootCmd.AddCommand(NewTerminateCommand())
	pluginRootCmd.AddCommand(cmd.NewVersionCmd(CLIName))

	addKubectlFlagsToCmd(pluginRootCmd)

	return pluginRootCmd
}

func addKubectlFlagsToCmd(cmd *cobra.Command) {
	// The "usual" clientcmd/kubectl flags
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	// cmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)
}

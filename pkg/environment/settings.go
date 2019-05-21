package environment

import (
	restclient "k8s.io/client-go/rest"

	"github.com/spf13/cobra"
)

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipctl is being used
	// out-of-cluster
	KubeConfigFilePath string

	// KubeConfig contains the configuration details needed for interacting
	// with the cluster
	KubeConfig *restclient.Config

	// Namespace is the kubernetes namespace to be used during the context of this action
	Namespace string

	// Debug is used for verbose output
	Debug bool
}

// InitFlags adds the default settings flags to cmd
func (a *AirshipCTLSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(&a.Debug, "debug", false, "enable verbose output")
	flags.StringVar(&a.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	flags.StringVar(&a.Namespace, "namespace", "default", "kubernetes namespace to use for the context of this command")
}

package environment

import (
	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
)

// Settings is a container for all of the settings needed by workflows
type Settings struct {
	*environment.AirshipCTLSettings

	// Namespace is the kubernetes namespace to be used during the context of this action
	Namespace string

	// AllNamespaces denotes whether or not to use all namespaces. It will override the Namespace string
	AllNamespaces bool

	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipctl is being used
	// out-of-cluster
	KubeConfigFilePath string
}

// InitFlags adds the default settings flags to cmd
func (s *Settings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVar(&s.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	flags.StringVar(&s.Namespace, "namespace", "default", "kubernetes namespace to use for the context of this command")
	flags.BoolVar(&s.AllNamespaces, "all-namespaces", false, "use all kubernetes namespaces for the context of this command")
}

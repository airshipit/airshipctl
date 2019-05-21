package environment

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

	// KubeClient contains a kubernetes clientset
	KubeClient kubernetes.Interface

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

// InitDefaults assigns default values for any value that has not been previously set
func (a *AirshipCTLSettings) InitDefaults() error {
	if a.KubeConfigFilePath == "" {
		a.KubeConfigFilePath = clientcmd.RecommendedHomeFile
	}

	var err error
	if a.KubeConfig == nil {
		a.KubeConfig, err = clientcmd.BuildConfigFromFlags("", a.KubeConfigFilePath)
		if err != nil {
			return err
		}
	}

	if a.KubeClient == nil {
		a.KubeClient, err = kubernetes.NewForConfig(a.KubeConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

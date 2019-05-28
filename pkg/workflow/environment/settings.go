package environment

import (
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	argo "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned"

	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
)

// Settings is a container for all of the settings needed by workflows
type Settings struct {
	// Namespace is the kubernetes namespace to be used during the context of this action
	Namespace string

	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipctl is being used
	// out-of-cluster
	KubeConfigFilePath string

	// KubeClient is an instrument for interacting with a kubernetes cluster
	KubeClient kubernetes.Interface

	// ArgoClient is an instrument for interacting with Argo workflows
	ArgoClient argo.Interface

	CRDClient apixv1beta1.ApiextensionsV1beta1Interface
}

// InitFlags adds the default settings flags to cmd
func (s *Settings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.StringVar(&s.KubeConfigFilePath, "kubeconfig", "", "path to kubeconfig")
	flags.StringVar(&s.Namespace, "namespace", "default", "kubernetes namespace to use for the context of this command")
}

// Init assigns default values
func (s *Settings) Init() error {
	if s.KubeConfigFilePath == "" {
		s.KubeConfigFilePath = clientcmd.RecommendedHomeFile
	}

	var err error
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", s.KubeConfigFilePath)
	if err != nil {
		return err
	}

	s.KubeClient, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}

	s.ArgoClient, err = argo.NewForConfig(kubeConfig)
	if err != nil {
		return err
	}
	return nil
}

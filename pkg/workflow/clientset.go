package workflow

import (
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apixv1beta1fake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd"

	argo "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned"
	argofake "github.com/ian-howell/airshipctl/pkg/client/clientset/versioned/fake"
	"github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

// Clientset is a container for the various clients that are useful to the workflow command
type Clientset struct {
	// Kube is an instrument for interacting with a kubernetes cluster
	Kube kubernetes.Interface

	// Argo is an instrument for interacting with Argo workflows
	Argo argo.Interface

	// CRD is an instrument for interacting with CRDs
	CRD apixv1beta1.Interface
}

var (
	clientset *Clientset
)

// GetClientset provides access to the clientset singleton
func GetClientset(settings *environment.Settings) (*Clientset, error) {
	if clientset != nil {
		return clientset, nil
	}

	if settings.KubeConfigFilePath == "" {
		settings.KubeConfigFilePath = clientcmd.RecommendedHomeFile
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", settings.KubeConfigFilePath)
	if err != nil {
		return nil, err
	}

	clientset = &Clientset{}
	clientset.Kube, err = kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	clientset.Argo, err = argo.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	clientset.CRD, err = apixv1beta1.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// NewSimpleClientset sets the singleton to a fake Clientset. It then returns the Clientset
func NewSimpleClientset(kubeObjs, argoObjs, crdObjs []runtime.Object) *Clientset {
	clientset = &Clientset{
		Kube: kubefake.NewSimpleClientset(kubeObjs...),
		Argo: argofake.NewSimpleClientset(argoObjs...),
		CRD:  apixv1beta1fake.NewSimpleClientset(crdObjs...),
	}
	return clientset
}

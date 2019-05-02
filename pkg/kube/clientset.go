package kubernetes

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// NewForConfig creates a kubernetes client using the config at $HOME/.kube/config
func NewForConfig(kubeconfigFilepath string) (kubernetes.Interface, error) {
	if kubeconfigFilepath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err.Error())
		}
		kubeconfigFilepath = filepath.Join(home, ".kube", "config")
	}

	// use the current context in kubeconfigFilepath
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilepath)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset, nil
}

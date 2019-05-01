package kubernetes

import (
	"flag"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient creates a kubernetes client using the config at $HOME/.kube/config
func GetClient() *kubernetes.Clientset {
	var kubeconfig *string
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	// TODO(howell): This was example code. The flag parsing needs to be
	// moved to a command
	fp := filepath.Join(home, ".kube", "config")
	kubeconfig = flag.String("kubeconfig", fp, "(optional) absolute path to the kubeconfig file")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

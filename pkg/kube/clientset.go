package kube

import (
	"errors"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ian-howell/airshipadm/pkg/util"
)

// Client is a device which communicates with the Kubernetes API
type Client struct {
	kubernetes.Interface
}

// NewForConfig creates a kubernetes client using the config at $HOME/.kube/config
func NewForConfig(kubeconfigFilepath string) (*Client, error) {
	if kubeconfigFilepath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.New("could not find kubernetes config file: " + err.Error())
		}
		kubeconfigFilepath = filepath.Join(home, ".kube", "config")
	}

	if err := util.IsReadable(kubeconfigFilepath); err != nil {
		return nil, errors.New("could not open " + kubeconfigFilepath + ": " + err.Error())
	}

	// use the current context in kubeconfigFilepath
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFilepath)
	if err != nil {
		return nil, errors.New("could not build kubernetes config: " + err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{clientset}, nil
}

package client

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
)

const (
	// buffDir is a directory that is used as a tmp storage for kubectl
	buffDir = ".airship"
)

// Interface provides an abstraction layer to interactions
// with kubernetes clusters by getting Clientset which includes
// all kubernetes core objects with standard operations and kubectl
// interface that is built on top of kubectl libraries and implements
// such kubectl subcommands as kubectl apply (more will be added)
type Interface interface {
	ClientSet() kubernetes.Interface
	Kubectl() kubectl.Interface
}

// Client is implementation of Cluster interface
type Client struct {
	kubectl   kubectl.Interface
	clientset kubernetes.Interface
}

// ClientSet getter for Clientset interface
func (c *Client) ClientSet() kubernetes.Interface {
	return c.clientset
}

// Kubectl getter for Kubectl interface
func (c *Client) Kubectl() kubectl.Interface {
	return c.kubectl
}

// NewClient returns Cluster interface with Kubectl
// and Clientset interfaces initialized
func NewClient(as *environment.AirshipCTLSettings) (Interface, error) {
	f := k8sutils.FactoryFromKubeconfigPath(as.KubeConfigPath())
	kctl := kubectl.NewKubectl(f).
		WithBufferDir(filepath.Dir(as.AirshipConfigPath()) + buffDir)
	clientSet, err := f.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	client := &Client{}
	client.SetClientset(clientSet)
	client.SetKubectl(kctl)
	return client, nil
}

// SetClientset setter for Clientset interface
func (c *Client) SetClientset(cs kubernetes.Interface) {
	c.clientset = cs
}

// SetKubectl setter for Kubectl interface
func (c *Client) SetKubectl(kctl kubectl.Interface) {
	c.kubectl = kctl
}

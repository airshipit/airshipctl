package client

import (
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
)

const (
	// buffDir is a directory that is used as a tmp storage for kubectl
	buffDir = ".airship"
)

// Interface provides an abstraction layer to interactions with kubernetes
// clusters by providing a ClientSet which includes all kubernetes core objects
// with standard operations, a DynamicClient which provides interactions with
// loosely typed kubernetes resources, and a Kubectl interface that is built on
// top of kubectl libraries and implements such kubectl subcommands as kubectl
// apply (more will be added)
type Interface interface {
	ClientSet() kubernetes.Interface
	DynamicClient() dynamic.Interface

	Kubectl() kubectl.Interface
}

// Client is an implementation of Interface
type Client struct {
	clientSet     kubernetes.Interface
	dynamicClient dynamic.Interface

	kubectl kubectl.Interface
}

// Client implements Interface
var _ Interface = &Client{}

// NewClient returns Cluster interface with Kubectl
// and ClientSet interfaces initialized
func NewClient(settings *environment.AirshipCTLSettings) (Interface, error) {
	client := new(Client)
	var err error

	f := k8sutils.FactoryFromKubeConfigPath(settings.KubeConfigPath())

	pathToBufferDir := filepath.Join(filepath.Dir(settings.AirshipConfigPath()), buffDir)
	client.kubectl = kubectl.NewKubectl(f).WithBufferDir(pathToBufferDir)

	client.clientSet, err = f.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	client.dynamicClient, err = f.DynamicClient()
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ClientSet getter for ClientSet interface
func (c *Client) ClientSet() kubernetes.Interface {
	return c.clientSet
}

// SetClientSet setter for ClientSet interface
func (c *Client) SetClientSet(clientSet kubernetes.Interface) {
	c.clientSet = clientSet
}

// DynamicClient getter for DynamicClient interface
func (c *Client) DynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// SetDynamicClient setter for DynamicClient interface
func (c *Client) SetDynamicClient(dynamicClient dynamic.Interface) {
	c.dynamicClient = dynamicClient
}

// Kubectl getter for Kubectl interface
func (c *Client) Kubectl() kubectl.Interface {
	return c.kubectl
}

// SetKubectl setter for Kubectl interface
func (c *Client) SetKubectl(kctl kubectl.Interface) {
	c.kubectl = kctl
}

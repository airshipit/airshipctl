/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package client

import (
	"path/filepath"

	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	k8sutils "opendev.org/airship/airshipctl/pkg/k8s/utils"
)

// Interface provides an abstraction layer to interactions with kubernetes
// clusters by providing the following:
// * A ClientSet which includes all kubernetes core objects with standard operations
// * A DynamicClient which provides interactions with loosely typed kubernetes resources
// * An ApiextensionsClientSet which provides interactions with CustomResourceDefinitions
// * A Kubectl interface that is built on top of kubectl libraries and
//   implements such kubectl subcommands as kubectl apply (more will be added)
type Interface interface {
	ClientSet() kubernetes.Interface
	DynamicClient() dynamic.Interface
	ApiextensionsClientSet() apix.Interface

	Kubectl() kubectl.Interface
}

// Client is an implementation of Interface
type Client struct {
	clientSet     kubernetes.Interface
	dynamicClient dynamic.Interface
	apixClient    apix.Interface

	kubectl kubectl.Interface
}

// Client implements Interface
var _ Interface = &Client{}

// Factory is a function which creates Interfaces
type Factory func(airshipConfigPath string, kubeconfig string) (Interface, error)

// DefaultClient is a factory which generates a default client
var DefaultClient Factory = NewClient

// NewClient creates a Client initialized from the passed in settings
func NewClient(airshipConfigPath string, kubeconfig string) (Interface, error) {
	client := new(Client)
	var err error

	// TODO add support for kubeconfig context, for now use current context
	f := k8sutils.FactoryFromKubeConfig(kubeconfig, "")

	pathToBufferDir := filepath.Dir(airshipConfigPath)
	client.kubectl = kubectl.NewKubectl(f).WithBufferDir(pathToBufferDir)

	client.clientSet, err = f.KubernetesClientSet()
	if err != nil {
		return nil, err
	}

	client.dynamicClient, err = f.DynamicClient()
	if err != nil {
		return nil, err
	}

	// kubectl factories can't create CRD clients...
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client.apixClient, err = apix.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ClientSet returns the ClientSet interface
func (c *Client) ClientSet() kubernetes.Interface {
	return c.clientSet
}

// SetClientSet sets the ClientSet interface
func (c *Client) SetClientSet(clientSet kubernetes.Interface) {
	c.clientSet = clientSet
}

// DynamicClient returns the DynamicClient interface
func (c *Client) DynamicClient() dynamic.Interface {
	return c.dynamicClient
}

// SetDynamicClient sets the DynamicClient interface
func (c *Client) SetDynamicClient(dynamicClient dynamic.Interface) {
	c.dynamicClient = dynamicClient
}

// ApiextensionsClientSet returns the Apiextensions interface
func (c *Client) ApiextensionsClientSet() apix.Interface {
	return c.apixClient
}

// SetApiextensionsClientSet sets the ApiextensionsClientSet interface
func (c *Client) SetApiextensionsClientSet(apixClient apix.Interface) {
	c.apixClient = apixClient
}

// Kubectl returns the Kubectl interface
func (c *Client) Kubectl() kubectl.Interface {
	return c.kubectl
}

// SetKubectl sets the Kubectl interface
func (c *Client) SetKubectl(kctl kubectl.Interface) {
	c.kubectl = kctl
}

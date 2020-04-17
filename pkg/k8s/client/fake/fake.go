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

package fake

import (
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apixFake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	kubernetesFake "k8s.io/client-go/kubernetes/fake"

	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
	"opendev.org/airship/airshipctl/testutil/k8sutils"
)

// Client is an implementation of client.Interface meant for testing purposes.
type Client struct {
	mockClientSet              func() kubernetes.Interface
	mockDynamicClient          func() dynamic.Interface
	mockApiextensionsClientSet func() apix.Interface
	mockKubectl                func() kubectl.Interface
}

var _ client.Interface = &Client{}

// ClientSet is used to get a mocked implementation of a kubernetes clientset.
// To initialize the mocked clientset to be returned, use the WithTypedObjects
// ResourceAccumulator
func (c *Client) ClientSet() kubernetes.Interface {
	return c.mockClientSet()
}

// DynamicClient is used to get a mocked implementation of a dynamic client.
// To initialize the mocked client to be returned, use the WithDynamicObjects
// ResourceAccumulator.
func (c *Client) DynamicClient() dynamic.Interface {
	return c.mockDynamicClient()
}

// ApiextensionsClientSet is used to get a mocked implementation of an
// Apiextensions clientset. To initialize the mocked client to be returned,
// use the WithCRDs ResourceAccumulator
func (c *Client) ApiextensionsClientSet() apix.Interface {
	return c.mockApiextensionsClientSet()
}

// Kubectl is used to get a mocked implementation of a Kubectl client.
// To initialize the mocked client to be returned, use the WithKubectl ResourceAccumulator
func (c *Client) Kubectl() kubectl.Interface {
	return c.mockKubectl()
}

// A ResourceAccumulator is an option meant to be passed to NewClient.
// ResourceAccumulators can be mixed and matched to create a collection of
// mocked clients, each having their own fake objects.
type ResourceAccumulator func(*Client)

// NewClient creates an instance of a Client. If no arguments are passed, the
// returned Client will have fresh mocked kubernetes clients which will have no
// prior knowledge of any resources.
//
// If prior knowledge of resources is desirable, NewClient should receive an
// appropriate ResourceAccumulator initialized with the desired resources.
func NewClient(resourceAccumulators ...ResourceAccumulator) *Client {
	fakeClient := new(Client)
	for _, accumulator := range resourceAccumulators {
		accumulator(fakeClient)
	}

	if fakeClient.mockClientSet == nil {
		fakeClient.mockClientSet = func() kubernetes.Interface {
			return kubernetesFake.NewSimpleClientset()
		}
	}
	if fakeClient.mockDynamicClient == nil {
		fakeClient.mockDynamicClient = func() dynamic.Interface {
			return dynamicFake.NewSimpleDynamicClient(runtime.NewScheme())
		}
	}
	if fakeClient.mockApiextensionsClientSet == nil {
		fakeClient.mockApiextensionsClientSet = func() apix.Interface {
			return apixFake.NewSimpleClientset()
		}
	}
	if fakeClient.mockKubectl == nil {
		fakeClient.mockKubectl = func() kubectl.Interface {
			return kubectl.NewKubectl(k8sutils.NewMockKubectlFactory())
		}
	}
	return fakeClient
}

// WithTypedObjects returns a ResourceAccumulator with resources which would
// normally be accessible through a kubernetes ClientSet (e.g. Pods,
// Deployments, etc...).
func WithTypedObjects(objs ...runtime.Object) ResourceAccumulator {
	return func(c *Client) {
		c.mockClientSet = func() kubernetes.Interface {
			return kubernetesFake.NewSimpleClientset(objs...)
		}
	}
}

// WithCRDs returns a ResourceAccumulator with resources which would
// normally be accessible through a kubernetes ApiextensionsClientSet (e.g. CRDs).
func WithCRDs(objs ...runtime.Object) ResourceAccumulator {
	return func(c *Client) {
		c.mockApiextensionsClientSet = func() apix.Interface {
			return apixFake.NewSimpleClientset(objs...)
		}
	}
}

// WithDynamicObjects returns a ResourceAccumulator with resources which would
// normally be accessible through a kubernetes DynamicClient (e.g. unstructured.Unstructured).
func WithDynamicObjects(objs ...runtime.Object) ResourceAccumulator {
	return func(c *Client) {
		c.mockDynamicClient = func() dynamic.Interface {
			return dynamicFake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)
		}
	}
}

// WithKubectl returns a ResourceAccumulator with an instance of a kubectl.Interface.
func WithKubectl(kubectlInstance *kubectl.Kubectl) ResourceAccumulator {
	return func(c *Client) {
		c.mockKubectl = func() kubectl.Interface {
			return kubectlInstance
		}
	}
}

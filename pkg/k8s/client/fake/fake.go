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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/k8s/kubectl"
)

// Client is an implementation of client.Interface meant for testing purposes.
// Its member methods are intended to be implemented on a case-by-case basis
// per test. Examples of implementations can be found with each interface
// method.
type Client struct {
	MockClientSet              func() kubernetes.Interface
	MockDynamicClient          func() dynamic.Interface
	MockApiextensionsClientSet func() apix.Interface
	MockKubectl                func() kubectl.Interface
}

var _ client.Interface = &Client{}

// ClientSet is used to get a mocked implementation of a kubernetes clientset.
// To initialize the mocked clientset to be returned, the MockClientSet method
// must be implemented, ideally returning a k8s.io/client-go/kubernetes/fake.Clientset.
//
// Example:
//
// testClient := fake.Client {
// 	MockClientSet: func() kubernetes.Interface {
// 		return kubernetes_fake.NewSimpleClientset()
// 	},
// }
func (c Client) ClientSet() kubernetes.Interface {
	return c.MockClientSet()
}

// DynamicClient is used to get a mocked implementation of a dynamic client.
// To initialize the mocked client to be returned, the MockDynamicClient method
// must be implemented, ideally returning a k8s.io/client-go/dynamic/fake.FakeDynamicClient.
//
// Example:
// Here, scheme is a k8s.io/apimachinery/pkg/runtime.Scheme, possibly created
// via runtime.NewScheme()
//
// testClient := fake.Client {
// 	MockDynamicClient: func() dynamic.Interface {
// 		return dynamic_fake.NewSimpleDynamicClient(scheme)
// 	},
// }
func (c Client) DynamicClient() dynamic.Interface {
	return c.MockDynamicClient()
}

// ApiextensionsClientSet is used to get a mocked implementation of an
// Apiextensions clientset.  To initialize the mocked client to be returned,
// the MockApiextensionsClientSet method must be implemented, ideally returning a
// k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake.ClientSet.
//
// Example:
//
// testClient := fake.Client {
// 	MockApiextensionsClientSet: func() apix.Interface {
// 		return apix_fake.NewSimpleClientset()
// 	},
// }
func (c Client) ApiextensionsClientSet() apix.Interface {
	return c.MockApiextensionsClientSet()
}

// Kubectl is used to get a mocked implementation of a Kubectl client.
// To initialize the mocked client to be returned, the MockKubectl method
// must be implemented.
//
// Example: TODO(howell)
func (c Client) Kubectl() kubectl.Interface {
	return c.MockKubectl()
}

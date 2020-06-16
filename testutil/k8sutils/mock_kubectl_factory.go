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

package k8sutils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util/openapi"
	"k8s.io/kubectl/pkg/validation"
)

// MockKubectlFactory implements Factory interface for testing purposes.
type MockKubectlFactory struct {
	MockToDiscoveryClient     func() (discovery.CachedDiscoveryInterface, error)
	MockDynamicClient         func() (dynamic.Interface, error)
	MockOpenAPISchema         func() (openapi.Resources, error)
	MockValidator             func() (validation.Schema, error)
	MockToRESTMapper          func() (meta.RESTMapper, error)
	MockToRESTConfig          func() (*rest.Config, error)
	MockNewBuilder            func() *resource.Builder
	MockToRawKubeConfigLoader func() clientcmd.ClientConfig
	MockClientForMapping      func() (resource.RESTClient, error)
	KubeConfig                kubeconfig.Config
	genericclioptions.ConfigFlags
	cmdutil.Factory
}

// ToDiscoveryClient implements Factory interface
func (f *MockKubectlFactory) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return f.MockToDiscoveryClient()
}

// DynamicClient implements Factory interface
// Returns a mock dynamic client ready for use
func (f *MockKubectlFactory) DynamicClient() (dynamic.Interface, error) { return f.MockDynamicClient() }

// OpenAPISchema implements Factory interface
// Returns a mock openapi schema definition. Schema definition includes metadata and structural information about
// Kubernetes object definitions
func (f *MockKubectlFactory) OpenAPISchema() (openapi.Resources, error) { return f.MockOpenAPISchema() }

// Validator implements Factory interface
// Returns a mock schema that can validate objects stored on disk
func (f *MockKubectlFactory) Validator(bool) (validation.Schema, error) {
	return f.MockValidator()
}

// ToRESTMapper implements Factory interface
// Returns a mock RESTMapper
// RESTMapper allows clients to map resources to kind, and map kind and version to interfaces for manipulating
// those objects. It is primarily intended for consumers of Kubernetes compatible REST APIs
func (f *MockKubectlFactory) ToRESTMapper() (meta.RESTMapper, error) { return f.MockToRESTMapper() }

// ToRESTConfig implements Factory interface
// Returns a mock Config
// Config holds the common attributes that can be passed to a Kubernetes client on initialization
func (f *MockKubectlFactory) ToRESTConfig() (*rest.Config, error) { return f.MockToRESTConfig() }

// NewBuilder implements Factory interface
// Returns a mock object that assists in loading objects from both disk and the server
func (f *MockKubectlFactory) NewBuilder() *resource.Builder { return f.MockNewBuilder() }

// ToRawKubeConfigLoader implements Factory interface
func (f *MockKubectlFactory) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return f.MockToRawKubeConfigLoader()
}

// ClientForMapping implements Factory interface
// Returns a mock RESTClient for working with the specified RESTMapping or an error
func (f *MockKubectlFactory) ClientForMapping(*meta.RESTMapping) (resource.RESTClient, error) {
	return f.MockClientForMapping()
}

// WithToDiscoveryClientByError returns mock discovery client with its respective error
func (f *MockKubectlFactory) WithToDiscoveryClientByError(d discovery.CachedDiscoveryInterface,
	err error) *MockKubectlFactory {
	f.MockToDiscoveryClient = func() (discovery.CachedDiscoveryInterface, error) { return d, err }
	return f
}

// WithOpenAPISchemaByError returns mock openAPISchema with its respective error
func (f *MockKubectlFactory) WithOpenAPISchemaByError(r openapi.Resources, err error) *MockKubectlFactory {
	f.MockOpenAPISchema = func() (openapi.Resources, error) { return r, err }
	return f
}

// WithDynamicClientByError returns mock dynamic client with its respective error
func (f *MockKubectlFactory) WithDynamicClientByError(d dynamic.Interface, err error) *MockKubectlFactory {
	f.MockDynamicClient = func() (dynamic.Interface, error) { return d, err }
	return f
}

// WithValidatorByError returns mock validator with its respective error
func (f *MockKubectlFactory) WithValidatorByError(v validation.Schema, err error) *MockKubectlFactory {
	f.MockValidator = func() (validation.Schema, error) { return v, err }
	return f
}

// WithToRESTMapperByError returns mock RESTMapper with its respective error
func (f *MockKubectlFactory) WithToRESTMapperByError(r meta.RESTMapper, err error) *MockKubectlFactory {
	f.MockToRESTMapper = func() (meta.RESTMapper, error) { return r, err }
	return f
}

// WithToRESTConfigByError returns mock RESTConfig with its respective error
func (f *MockKubectlFactory) WithToRESTConfigByError(r *rest.Config, err error) *MockKubectlFactory {
	f.MockToRESTConfig = func() (*rest.Config, error) { return r, err }
	return f
}

// WithNewBuilderByError returns mock resource builder with its respective error
func (f *MockKubectlFactory) WithNewBuilderByError(r *resource.Builder) *MockKubectlFactory {
	f.MockNewBuilder = func() *resource.Builder { return r }
	return f
}

// WithToRawKubeConfigLoaderByError returns mock raw kubeconfig loader with its respective error
func (f *MockKubectlFactory) WithToRawKubeConfigLoaderByError(c clientcmd.ClientConfig) *MockKubectlFactory {
	f.MockToRawKubeConfigLoader = func() clientcmd.ClientConfig { return c }
	return f
}

// WithClientForMappingByError returns mock client mapping with its respective error
func (f *MockKubectlFactory) WithClientForMappingByError(r resource.RESTClient, err error) *MockKubectlFactory {
	f.MockClientForMapping = func() (resource.RESTClient, error) { return r, err }
	return f
}

// NewMockKubectlFactory defines the functions of MockKubectlFactory with nil values for testing purpose
func NewMockKubectlFactory() *MockKubectlFactory {
	return &MockKubectlFactory{MockDynamicClient: func() (dynamic.Interface, error) { return nil, nil },
		MockToDiscoveryClient:     func() (discovery.CachedDiscoveryInterface, error) { return nil, nil },
		MockOpenAPISchema:         func() (openapi.Resources, error) { return nil, nil },
		MockValidator:             func() (validation.Schema, error) { return nil, nil },
		MockToRESTMapper:          func() (meta.RESTMapper, error) { return nil, nil },
		MockToRESTConfig:          func() (*rest.Config, error) { return nil, nil },
		MockNewBuilder:            func() *resource.Builder { return nil },
		MockToRawKubeConfigLoader: func() clientcmd.ClientConfig { return nil },
		MockClientForMapping:      func() (resource.RESTClient, error) { return nil, nil },
	}
}

// MockClientConfig implements DirectClientConfig interface
// Returns mock client config for testing
type MockClientConfig struct {
	clientcmd.DirectClientConfig
	MockNamespace func() (string, bool, error)
}

// Namespace returns mock namespace for testing
func (c MockClientConfig) Namespace() (string, bool, error) { return c.MockNamespace() }

// WithNamespace returns mock namespace with its respective error
func (c *MockClientConfig) WithNamespace(s string, b bool, err error) *MockClientConfig {
	c.MockNamespace = func() (string, bool, error) { return s, b, err }
	return c
}

// NewMockClientConfig returns mock client config for testing
func NewMockClientConfig() *MockClientConfig {
	return &MockClientConfig{
		MockNamespace: func() (string, bool, error) { return "test", false, nil },
	}
}

// ClientHandler is an interface that can be injected into FakeFactory
// it's purpose to mock http request handling done by the Kubernetes Clients produced by cmdutils.Factory
type ClientHandler interface {
	Handle(t *testing.T, req *http.Request) (*http.Response, bool, error)
}

var (
	nsNamedPathRegex = regexp.MustCompile(`/api/v1/namespaces/([^/]+)`)
	nsPath           = "/api/v1/namespaces"
)

// NamespaceHandler implements ClientHandler, that is to be used to handle
// Http Requests made by clients that are produced by cmdutils.Factory interface
type NamespaceHandler struct {
}

var _ ClientHandler = &NamespaceHandler{}

// Handle implements handler
func (h *NamespaceHandler) Handle(_ *testing.T, req *http.Request) (*http.Response, bool, error) {
	c := scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
	switch match, method := nsNamedPathRegex.FindStringSubmatch(req.URL.Path), req.Method; {
	case match != nil && method == http.MethodGet:
		ns := &corev1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1"},
			ObjectMeta: v1.ObjectMeta{
				// check that this index exists is performed at case statement match != nil
				// this means that [0] and [1] exist
				Name: match[1],
			}}
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     cmdtesting.DefaultHeader(),
			Body:       cmdtesting.ObjBody(c, ns)}
		return response, true, nil

	case req.URL.Path == nsPath && method == http.MethodPost:
		ns := &corev1.Namespace{
			TypeMeta: v1.TypeMeta{
				Kind:       "Namespace",
				APIVersion: "v1"},
		}
		response := &http.Response{StatusCode: http.StatusOK,
			Header: cmdtesting.DefaultHeader(),
			Body:   cmdtesting.ObjBody(c, ns)}
		return response, true, nil
	}

	return nil, false, nil
}

// InventoryObjectHandler handles configmap inventory object from cli-utils by mocking
// http calls made by clients produced by cmdutils.Factory interface
type InventoryObjectHandler struct {
	inventoryObj *corev1.ConfigMap
}

var _ ClientHandler = &InventoryObjectHandler{}

var (
	cmPathRegex              = regexp.MustCompile(`^/namespaces/([^/]+)/configmaps$`)
	resourceNameRegexpString = `^[a-zA-Z]+-[a-z0-9]+$`
	invObjNameRegex          = regexp.MustCompile(resourceNameRegexpString)
	invObjPathRegex          = regexp.MustCompile(`^/namespaces/([^/]+)/configmaps/` + resourceNameRegexpString[:1])
	codec                    = scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...)
)

// Handle implements handler
func (i *InventoryObjectHandler) Handle(t *testing.T, req *http.Request) (*http.Response, bool, error) {
	if req.Method == http.MethodPost && cmPathRegex.Match([]byte(req.URL.Path)) {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, false, err
		}
		cm := corev1.ConfigMap{}
		err = runtime.DecodeInto(codec, b, &cm)
		if err != nil {
			return nil, false, err
		}
		if invObjNameRegex.Match([]byte(cm.Name)) {
			i.inventoryObj = &cm
			bodyRC := ioutil.NopCloser(bytes.NewReader(b))
			return &http.Response{StatusCode: http.StatusCreated, Header: cmdtesting.DefaultHeader(), Body: bodyRC}, true, nil
		}
		return nil, false, nil
	}

	if req.Method == http.MethodGet && cmPathRegex.Match([]byte(req.URL.Path)) {
		cmList := corev1.ConfigMapList{
			TypeMeta: v1.TypeMeta{
				APIVersion: "v1",
				Kind:       "List",
			},
			Items: []corev1.ConfigMap{},
		}
		if i.inventoryObj != nil {
			cmList.Items = append(cmList.Items, *i.inventoryObj)
		}
		bodyRC := ioutil.NopCloser(bytes.NewReader(toJSONBytes(t, &cmList)))
		return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: bodyRC}, true, nil
	}

	if req.Method == http.MethodGet && invObjPathRegex.Match([]byte(req.URL.Path)) {
		if i.inventoryObj == nil {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Header:     cmdtesting.DefaultHeader(),
				Body:       cmdtesting.StringBody("")}, true, nil
		}
		bodyRC := ioutil.NopCloser(bytes.NewReader(toJSONBytes(t, i.inventoryObj)))
		return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: bodyRC}, true, nil
	}
	return nil, false, nil
}

// GenericHandler is a handler for generic objects
type GenericHandler struct {
	Obj       runtime.Object
	Namespace string
	// URLPath is a string for formater in which it should be defined how to inject a namespace into it
	// example : /namespaces/%s/deployments
	URLPath string
	Bytes   []byte
}

var _ ClientHandler = &GenericHandler{}

// Handle implements handler
func (g *GenericHandler) Handle(t *testing.T, req *http.Request) (*http.Response, bool, error) {
	err := runtime.DecodeInto(codec, g.Bytes, g.Obj)
	if err != nil {
		return nil, false, err
	}
	accessor, err := meta.Accessor(g.Obj)
	if err != nil {
		return nil, false, err
	}
	basePath := fmt.Sprintf(g.URLPath, g.Namespace)
	resourcePath := path.Join(basePath, accessor.GetName())
	if req.URL.Path == resourcePath && req.Method == http.MethodGet {
		bodyRC := ioutil.NopCloser(bytes.NewReader(toJSONBytes(t, g.Obj)))
		return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: bodyRC}, true, nil
	}
	if req.URL.Path == resourcePath && req.Method == http.MethodPatch {
		bodyRC := ioutil.NopCloser(bytes.NewReader(toJSONBytes(t, g.Obj)))
		return &http.Response{StatusCode: http.StatusOK, Header: cmdtesting.DefaultHeader(), Body: bodyRC}, true, nil
	}
	return nil, false, nil
}

func toJSONBytes(t *testing.T, obj runtime.Object) []byte {
	objBytes, err := runtime.Encode(unstructured.NewJSONFallbackEncoder(codec), obj)
	require.NoError(t, err)
	return objBytes
}

// FakeFactory returns a fake factory based on provided handlers
func FakeFactory(t *testing.T, handlers []ClientHandler) *cmdtesting.TestFactory {
	f := cmdtesting.NewTestFactory().WithNamespace("test")
	defer f.Cleanup()
	testRESTClient := &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			for _, h := range handlers {
				resp, handled, err := h.Handle(t, req)
				if handled {
					return resp, err
				}
			}
			t.Fatalf("unexpected request: %#v\n%#v", req.URL, req)
			// dummy return
			return nil, nil
		}),
	}
	f.Client = testRESTClient
	f.UnstructuredClient = testRESTClient
	return f
}

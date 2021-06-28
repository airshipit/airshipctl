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

package poller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/openapi"
	"k8s.io/kubectl/pkg/validation"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/engine"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/event"
	kstatustestutil "sigs.k8s.io/cli-utils/pkg/kstatus/polling/testutil"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"opendev.org/airship/airshipctl/pkg/k8s/poller"
)

func TestNewStatusPoller(t *testing.T) {
	testCases := map[string]struct {
		factory       cmdutil.Factory
		expectedError bool
	}{
		"failed rest config": {
			factory: &MockCmdUtilFactory{MockToRESTConfig: func() (*rest.Config, error) {
				return nil, errors.New("rest config error")
			}},
			expectedError: true,
		},
		"failed rest mapper": {
			factory: &MockCmdUtilFactory{MockToRESTConfig: func() (*rest.Config, error) {
				return nil, nil
			},
				MockToRESTMapper: func() (meta.RESTMapper, error) {
					return nil, errors.New("rest mapper error")
				}},
			expectedError: true,
		},
		"failed new client": {
			factory: &MockCmdUtilFactory{MockToRESTConfig: func() (*rest.Config, error) {
				return nil, nil
			},
				MockToRESTMapper: func() (meta.RESTMapper, error) {
					return nil, nil
				}},
			expectedError: true,
		},
		"success new poller": {
			factory: &MockCmdUtilFactory{MockToRESTConfig: func() (*rest.Config, error) {
				return &rest.Config{}, nil
			},
				MockToRESTMapper: func() (meta.RESTMapper, error) {
					return testutil.NewFakeRESTMapper(), nil
				}},
			expectedError: false,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			p, err := poller.NewStatusPoller(tc.factory)
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, p)
			}
		})
	}
}

func TestStatusPollerRun(t *testing.T) {
	testCases := map[string]struct {
		identifiers              []object.ObjMetadata
		ClusterReaderFactoryFunc engine.ClusterReaderFactoryFunc
		StatusReadersFactoryFunc engine.StatusReadersFactoryFunc
		defaultStatusReader      engine.StatusReader
		expectedEventTypes       []event.EventType
	}{
		"single resource": {
			identifiers: []object.ObjMetadata{
				{
					GroupKind: schema.GroupKind{
						Group: "apps",
						Kind:  "Deployment",
					},
					Name:      "foo",
					Namespace: "bar",
				},
			},
			defaultStatusReader: &fakeStatusReader{
				resourceStatuses: map[schema.GroupKind][]status.Status{
					schema.GroupKind{Group: "apps", Kind: "Deployment"}: { //nolint:gofmt
						status.InProgressStatus,
						status.CurrentStatus,
					},
				},
				resourceStatusCount: make(map[schema.GroupKind]int),
			},
			expectedEventTypes: []event.EventType{
				event.ResourceUpdateEvent,
				event.ResourceUpdateEvent,
			},
			ClusterReaderFactoryFunc: func(_ client.Reader, _ meta.RESTMapper, _ []object.ObjMetadata) (
				engine.ClusterReader, error) {
				return kstatustestutil.NewNoopClusterReader(), nil
			},
			StatusReadersFactoryFunc: func(_ engine.ClusterReader, _ meta.RESTMapper) (
				statusReaders map[schema.GroupKind]engine.StatusReader, defaultStatusReader engine.StatusReader) {
				return make(map[schema.GroupKind]engine.StatusReader), &fakeStatusReader{
					resourceStatuses: map[schema.GroupKind][]status.Status{
						schema.GroupKind{Group: "apps", Kind: "Deployment"}: { //nolint:gofmt
							status.InProgressStatus,
							status.CurrentStatus,
						},
					},
					resourceStatusCount: make(map[schema.GroupKind]int),
				}
			},
		},
		"multiple resources": {
			identifiers: []object.ObjMetadata{
				{
					GroupKind: schema.GroupKind{
						Group: "apps",
						Kind:  "Deployment",
					},
					Name:      "foo",
					Namespace: "default",
				},
				{
					GroupKind: schema.GroupKind{
						Group: "",
						Kind:  "Service",
					},
					Name:      "bar",
					Namespace: "default",
				},
			},
			ClusterReaderFactoryFunc: func(_ client.Reader, _ meta.RESTMapper, _ []object.ObjMetadata) (
				engine.ClusterReader, error) {
				return kstatustestutil.NewNoopClusterReader(), nil
			},
			StatusReadersFactoryFunc: func(_ engine.ClusterReader, _ meta.RESTMapper) (
				statusReaders map[schema.GroupKind]engine.StatusReader, defaultStatusReader engine.StatusReader) {
				return make(map[schema.GroupKind]engine.StatusReader), &fakeStatusReader{
					resourceStatuses: map[schema.GroupKind][]status.Status{
						schema.GroupKind{Group: "apps", Kind: "Deployment"}: { //nolint:gofmt
							status.InProgressStatus,
							status.CurrentStatus,
						},
						schema.GroupKind{Group: "", Kind: "Service"}: { //nolint:gofmt
							status.InProgressStatus,
							status.InProgressStatus,
							status.CurrentStatus,
						},
					},
					resourceStatusCount: make(map[schema.GroupKind]int),
				}
			},
			expectedEventTypes: []event.EventType{
				event.ResourceUpdateEvent,
				event.ResourceUpdateEvent,
				event.ResourceUpdateEvent,
				event.ResourceUpdateEvent,
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			identifiers := tc.identifiers

			fakeMapper := testutil.NewFakeRESTMapper(
				appsv1.SchemeGroupVersion.WithKind("Deployment"),
				v1.SchemeGroupVersion.WithKind("Service"),
			)

			e := poller.StatusPoller{
				ClusterReaderFactoryFunc: tc.ClusterReaderFactoryFunc,
				StatusReadersFactoryFunc: tc.StatusReadersFactoryFunc,
				Engine:                   &engine.PollerEngine{Mapper: fakeMapper},
			}

			options := polling.Options{PollInterval: time.Second, UseCache: true}

			eventChannel := e.Poll(ctx, identifiers, options)

			var eventTypes []event.EventType
			for ch := range eventChannel {
				eventTypes = append(eventTypes, ch.EventType)
				if len(eventTypes) == len(tc.expectedEventTypes) {
					cancel()
				}
			}

			require.Equal(t, tc.expectedEventTypes, eventTypes)
		})
	}
}

type fakeStatusReader struct {
	resourceStatuses    map[schema.GroupKind][]status.Status
	resourceStatusCount map[schema.GroupKind]int
}

func (f *fakeStatusReader) ReadStatus(_ context.Context, identifier object.ObjMetadata) *event.ResourceStatus {
	count := f.resourceStatusCount[identifier.GroupKind]
	resourceStatusSlice := f.resourceStatuses[identifier.GroupKind]
	var resourceStatus status.Status
	if len(resourceStatusSlice) > count {
		resourceStatus = resourceStatusSlice[count]
	} else {
		resourceStatus = resourceStatusSlice[len(resourceStatusSlice)-1]
	}
	f.resourceStatusCount[identifier.GroupKind] = count + 1
	return &event.ResourceStatus{
		Identifier: identifier,
		Status:     resourceStatus,
	}
}

func (f *fakeStatusReader) ReadStatusForObject(_ context.Context, _ *unstructured.Unstructured) *event.ResourceStatus {
	return nil
}

var _ cmdutil.Factory = &MockCmdUtilFactory{}

type MockCmdUtilFactory struct {
	MockToRESTConfig func() (*rest.Config, error)
	MockToRESTMapper func() (meta.RESTMapper, error)
}

func (n *MockCmdUtilFactory) ToRESTConfig() (*rest.Config, error) {
	return n.MockToRESTConfig()
}

func (n *MockCmdUtilFactory) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) ToRESTMapper() (meta.RESTMapper, error) {
	return n.MockToRESTMapper()
}

func (n *MockCmdUtilFactory) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

func (n *MockCmdUtilFactory) DynamicClient() (dynamic.Interface, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) KubernetesClientSet() (*kubernetes.Clientset, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) RESTClient() (*rest.RESTClient, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) NewBuilder() *resource.Builder {
	return nil
}

func (n *MockCmdUtilFactory) ClientForMapping(_ *meta.RESTMapping) (resource.RESTClient, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) UnstructuredClientForMapping(_ *meta.RESTMapping) (resource.RESTClient, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) Validator(_ bool) (validation.Schema, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) OpenAPISchema() (openapi.Resources, error) {
	return nil, nil
}

func (n *MockCmdUtilFactory) OpenAPIGetter() discovery.OpenAPISchemaInterface {
	return nil
}

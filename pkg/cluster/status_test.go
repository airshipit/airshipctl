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

package cluster_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicFake "k8s.io/client-go/dynamic/fake"

	"opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

type fakeBundle struct {
	document.Bundle

	mockGetByGvk func(string, string, string) ([]document.Document, error)
}

func (fb fakeBundle) GetByGvk(group, version, kind string) ([]document.Document, error) {
	return fb.mockGetByGvk(group, version, kind)
}

func TestNewStatusMapErrorCases(t *testing.T) {
	dummyError := errors.New("test error")
	tests := []struct {
		name   string
		bundle document.Bundle
		err    error
	}{
		{
			name: "bundle-fails-retrieving-v1-resources",
			bundle: fakeBundle{
				mockGetByGvk: func(_, version, _ string) ([]document.Document, error) {
					if version == "v1" {
						return nil, dummyError
					}
					return nil, nil
				},
			},
			err: dummyError,
		},
		{
			name: "bundle-fails-retrieving-v1beta1-resources",
			bundle: fakeBundle{
				mockGetByGvk: func(_, version, _ string) ([]document.Document, error) {
					if version == "v1beta1" {
						return nil, dummyError
					}
					return nil, nil
				},
			},
			err: dummyError,
		},
		{
			name:   "no-failure-when-missing-status-check-annotation",
			bundle: testutil.NewTestBundle(t, "testdata/missing-status-check"),
			err:    nil,
		},
		{
			name:   "missing-status",
			bundle: testutil.NewTestBundle(t, "testdata/missing-status"),
			err:    cluster.ErrInvalidStatusCheck{What: "missing status field"},
		},
		{
			name:   "missing-condition",
			bundle: testutil.NewTestBundle(t, "testdata/missing-condition"),
			err:    cluster.ErrInvalidStatusCheck{What: "missing condition field"},
		},
		{
			name:   "malformed-status-check",
			bundle: testutil.NewTestBundle(t, "testdata/malformed-status-check"),
			err: cluster.ErrInvalidStatusCheck{What: `unable to parse jsonpath: ` +
				`"{invalid json": invalid character 'i' looking for beginning of object key string`},
		},
	}

	for _, tt := range tests {
		tt := tt
		_, err := cluster.NewStatusMap(tt.bundle)
		assert.Equal(t, tt.err, err)
	}
}

func TestGetStatusForResource(t *testing.T) {
	tests := []struct {
		name           string
		selector       document.Selector
		testClient     fake.Client
		expectedStatus cluster.Status
		err            error
	}{
		{
			name: "stable-resource-is-stable",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("stable-resource"),
			testClient:     makeTestClient(makeResource("Resource", "stable-resource", "stable")),
			expectedStatus: cluster.Status("Stable"),
		},
		{
			name: "pending-resource-is-pending",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("pending-resource"),
			testClient:     makeTestClient(makeResource("Resource", "pending-resource", "pending")),
			expectedStatus: cluster.Status("Pending"),
		},
		{
			name: "unknown-resource-is-unknown",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Resource").
				ByName("unknown"),
			testClient:     makeTestClient(makeResource("Resource", "unknown", "unknown")),
			expectedStatus: cluster.UnknownStatus,
		},
		{
			name: "stable-legacy-is-stable",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Legacy").
				ByName("stable-legacy"),
			testClient:     makeTestClient(makeResource("Legacy", "stable-legacy", "stable")),
			expectedStatus: cluster.Status("Stable"),
		},
		{
			name: "missing-resource-returns-error",
			selector: document.NewSelector().
				ByGvk("example.com", "v1", "Missing").
				ByName("missing-resource"),
			testClient: makeTestClient(),
			err:        cluster.ErrResourceNotFound{Resource: "missing-resource"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			bundle := testutil.NewTestBundle(t, "testdata/statusmap")
			testStatusMap, err := cluster.NewStatusMap(bundle)
			require.NoError(t, err)

			doc, err := bundle.SelectOne(tt.selector)
			require.NoError(t, err)

			actualStatus, err := testStatusMap.GetStatusForResource(tt.testClient, doc)
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
				// We expected an error - no need to check anything else
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, actualStatus)
		})
	}
}

func makeTestClient(obj ...runtime.Object) fake.Client {
	testClient := fake.Client{
		MockDynamicClient: func() dynamic.Interface {
			return dynamicFake.NewSimpleDynamicClient(runtime.NewScheme(), obj...)
		},
	}
	return testClient
}

func makeResource(kind, name, state string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "example.com/v1",
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"state": state,
			},
		},
	}
}

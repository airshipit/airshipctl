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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/testutil"
	"sigs.k8s.io/cli-utils/pkg/kstatus/status"
	"sigs.k8s.io/cli-utils/pkg/object"
	fakemapper "sigs.k8s.io/cli-utils/pkg/testutil"

	"opendev.org/airship/airshipctl/pkg/k8s/poller"
)

var (
	customGVK = schema.GroupVersionKind{
		Group:   "custom.io",
		Version: "v1beta1",
		Kind:    "Custom",
	}
	name      = "Foo"
	namespace = "default"
)

func TestGenericStatusReader(t *testing.T) {
	testCases := map[string]struct {
		result             *status.Result
		err                error
		expectedIdentifier object.ObjMetadata
		expectedStatus     status.Status
		condMap            map[schema.GroupKind]poller.Expression
	}{
		"successfully computes status": {
			result: &status.Result{
				Status:  status.InProgressStatus,
				Message: "this is a test",
			},
			expectedIdentifier: object.ObjMetadata{
				GroupKind: customGVK.GroupKind(),
				Name:      name,
				Namespace: namespace,
			},
			expectedStatus: status.InProgressStatus,
		},
		"successfully computes custom status": {
			result: &status.Result{
				Status:  status.CurrentStatus,
				Message: "this is a test",
			},
			expectedIdentifier: object.ObjMetadata{
				GroupKind: customGVK.GroupKind(),
				Name:      name,
				Namespace: namespace,
			},
			condMap: map[schema.GroupKind]poller.Expression{
				customGVK.GroupKind(): {Condition: "{.metadata.name}", Value: "Bar"}},
			expectedStatus: status.InProgressStatus,
		},
		"computing status fails": {
			err: fmt.Errorf("this error is a test"),
			expectedIdentifier: object.ObjMetadata{
				GroupKind: customGVK.GroupKind(),
				Name:      name,
				Namespace: namespace,
			},
			expectedStatus: status.UnknownStatus,
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			fakeReader := testutil.NewNoopClusterReader()
			fakeMapper := fakemapper.NewFakeRESTMapper()

			resourceStatusReader := &poller.CustomResourceReader{
				Reader: fakeReader,
				Mapper: fakeMapper,
				StatusFunc: func(u *unstructured.Unstructured) (*status.Result, error) {
					return tc.result, tc.err
				},
				CondMap: tc.condMap,
			}

			o := &unstructured.Unstructured{}
			o.SetGroupVersionKind(customGVK)
			o.SetName(name)
			o.SetNamespace(namespace)

			resourceStatus := resourceStatusReader.ReadStatusForObject(context.Background(), o)

			require.Equal(t, tc.expectedIdentifier, resourceStatus.Identifier)
			require.Equal(t, tc.expectedStatus, resourceStatus.Status)
		})
	}
}

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
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling/clusterreader"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/cli-utils/pkg/testutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"opendev.org/airship/airshipctl/pkg/k8s/poller"
)

var (
	deploymentGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")
	rsGVK         = appsv1.SchemeGroupVersion.WithKind("ReplicaSet")
	podGVK        = v1.SchemeGroupVersion.WithKind("Pod")
)

// gkNamespace contains information about a GroupVersionKind and a namespace.
type gkNamespace struct {
	GroupKind schema.GroupKind
	Namespace string
}

func TestSync(t *testing.T) {
	testCases := map[string]struct {
		identifiers    []object.ObjMetadata
		expectedSynced []gkNamespace
	}{
		"no identifiers": {
			identifiers: []object.ObjMetadata{},
		},
		"same GVK in multiple namespaces": {
			identifiers: []object.ObjMetadata{
				{
					GroupKind: deploymentGVK.GroupKind(),
					Name:      "deployment",
					Namespace: "Foo",
				},
				{
					GroupKind: deploymentGVK.GroupKind(),
					Name:      "deployment",
					Namespace: "Bar",
				},
			},
			expectedSynced: []gkNamespace{
				{
					GroupKind: deploymentGVK.GroupKind(),
					Namespace: "Foo",
				},
				{
					GroupKind: rsGVK.GroupKind(),
					Namespace: "Foo",
				},
				{
					GroupKind: podGVK.GroupKind(),
					Namespace: "Foo",
				},
				{
					GroupKind: deploymentGVK.GroupKind(),
					Namespace: "Bar",
				},
				{
					GroupKind: rsGVK.GroupKind(),
					Namespace: "Bar",
				},
				{
					GroupKind: podGVK.GroupKind(),
					Namespace: "Bar",
				},
			},
		},
	}

	fakeMapper := testutil.NewFakeRESTMapper(
		appsv1.SchemeGroupVersion.WithKind("Deployment"),
		appsv1.SchemeGroupVersion.WithKind("ReplicaSet"),
		v1.SchemeGroupVersion.WithKind("Pod"),
	)

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			fakeReader := &fakeReader{}

			cr, err := clusterreader.NewCachingClusterReader(fakeReader, fakeMapper, tc.identifiers)
			require.NoError(t, err)

			clusterReader := &poller.CachingClusterReader{Cr: cr}
			err = clusterReader.Sync(context.Background())
			require.NoError(t, err)

			synced := fakeReader.syncedGVKNamespaces
			sortGVKNamespaces(synced)
			expectedSynced := tc.expectedSynced
			sortGVKNamespaces(expectedSynced)
			require.Equal(t, expectedSynced, synced)
		})
	}
}

func TestSync_Errors(t *testing.T) {
	testCases := map[string]struct {
		mapper          meta.RESTMapper
		readerError     error
		expectSyncError bool
		cacheError      bool
		cacheErrorText  string
	}{
		"mapping and reader are successful": {
			mapper: testutil.NewFakeRESTMapper(
				apiextv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			),
			readerError:     nil,
			expectSyncError: false,
			cacheError:      false,
		},
		"reader returns NotFound error": {
			mapper: testutil.NewFakeRESTMapper(
				apiextv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			),
			readerError: errors.NewNotFound(schema.GroupResource{
				Group:    "apiextensions.k8s.io",
				Resource: "customresourcedefinitions",
			}, "my-crd"),
			expectSyncError: false,
			cacheError:      true,
			cacheErrorText:  "not found",
		},
		"reader returns request timed out other error": {
			mapper: testutil.NewFakeRESTMapper(
				apiextv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			),
			readerError:     errors.NewInternalError(fmt.Errorf("request timed out")),
			expectSyncError: false,
			cacheError:      true,
			cacheErrorText:  "not found",
		},
		"reader returns other error": {
			mapper: testutil.NewFakeRESTMapper(
				apiextv1.SchemeGroupVersion.WithKind("CustomResourceDefinition"),
			),
			readerError:     errors.NewInternalError(fmt.Errorf("testing")),
			expectSyncError: true,
			cacheError:      true,
			cacheErrorText:  "not found",
		},
		"mapping not found": {
			mapper:          testutil.NewFakeRESTMapper(),
			expectSyncError: false,
			cacheError:      true,
			cacheErrorText:  "no matches for kind",
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			identifiers := []object.ObjMetadata{
				{
					Name: "my-crd",
					GroupKind: schema.GroupKind{
						Group: "apiextensions.k8s.io",
						Kind:  "CustomResourceDefinition",
					},
				},
			}

			fakeReader := &fakeReader{
				err: tc.readerError,
			}

			cr, err := clusterreader.NewCachingClusterReader(fakeReader, tc.mapper, identifiers)
			require.NoError(t, err)
			clusterReader := poller.CachingClusterReader{Cr: cr}

			err = clusterReader.Sync(context.Background())

			if tc.expectSyncError {
				require.Equal(t, tc.readerError, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func sortGVKNamespaces(gvkNamespaces []gkNamespace) {
	sort.Slice(gvkNamespaces, func(i, j int) bool {
		if gvkNamespaces[i].GroupKind.String() != gvkNamespaces[j].GroupKind.String() {
			return gvkNamespaces[i].GroupKind.String() < gvkNamespaces[j].GroupKind.String()
		}
		return gvkNamespaces[i].Namespace < gvkNamespaces[j].Namespace
	})
}

type fakeReader struct {
	syncedGVKNamespaces []gkNamespace
	err                 error
}

func (f *fakeReader) Get(_ context.Context, _ client.ObjectKey, _ client.Object) error {
	return nil
}

//nolint:gocritic
func (f *fakeReader) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	var namespace string
	for _, opt := range opts {
		switch opt := opt.(type) {
		case client.InNamespace:
			namespace = string(opt)
		}
	}

	gvk := list.GetObjectKind().GroupVersionKind()
	f.syncedGVKNamespaces = append(f.syncedGVKNamespaces, gkNamespace{
		GroupKind: gvk.GroupKind(),
		Namespace: namespace,
	})

	return f.err
}

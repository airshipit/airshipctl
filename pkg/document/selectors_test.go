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

package document_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"

	airapiv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/testutil"
)

func TestSelectorsPositive(t *testing.T) {
	bundle := testutil.NewTestBundle(t, "testdata/selectors/valid")

	t.Run("TestEphemeralCloudDataSelector", func(t *testing.T) {
		doc, err := bundle.Select(document.NewEphemeralCloudDataSelector())
		require.NoError(t, err)
		assert.Len(t, doc, 1)
	})

	t.Run("TestEphemeralNetworkDataSelector", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		bmhDoc := docs[0]
		selector, err := document.NewNetworkDataSelector(bmhDoc)
		require.NoError(t, err)
		assert.Equal(t, "validName", selector.Name)
	})

	t.Run("TestEphemeralCloudDataSelector", func(t *testing.T) {
		doc, err := bundle.Select(document.NewEphemeralCloudDataSelector())
		require.NoError(t, err)
		assert.Len(t, doc, 1)
	})

	t.Run("TestNewClusterctlMetadataSelector", func(t *testing.T) {
		doc, err := bundle.Select(document.NewClusterctlMetadataSelector())
		require.NoError(t, err)
		assert.Len(t, doc, 1)
	})
}

func TestSelectorsNegative(t *testing.T) {
	// These two tests take bundle with two malformed documents
	// each of the documents will fail at different locations providing higher
	// test coverage
	bundle := testutil.NewTestBundle(t, "testdata/selectors/invalid")

	t.Run("TestNewNetworkDataSelectorErr", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 2)
		bmhDoc := docs[0]
		_, err = document.NewNetworkDataSelector(bmhDoc)
		assert.Error(t, err)
	})

	t.Run("TestEphemeralNetworkDataSelectorErr", func(t *testing.T) {
		docs, err := bundle.Select(document.NewEphemeralBMHSelector())
		require.NoError(t, err)
		assert.Len(t, docs, 2)
		bmhDoc := docs[1]
		_, err = document.NewNetworkDataSelector(bmhDoc)
		assert.Error(t, err)
	})
}

func TestSelectorsSkip(t *testing.T) {
	// These two tests take bundle with two malformed documents
	// each of the documents will fail at different locations providing higher
	// test coverage
	bundle := testutil.NewTestBundle(t, "testdata/selectors/exclude-from-k8s")

	t.Run("TestNewNetworkDataSelectorErr", func(t *testing.T) {
		selector := document.NewDeployToK8sSelector()
		docs, err := bundle.Select(selector)
		require.NoError(t, err)
		assert.Len(t, docs, 5)
		for _, doc := range docs {
			assert.NotEqual(t, "ignore-namespace", doc.GetName())
			assert.NotEqual(t, "ignore-bmh", doc.GetName())
		}
	})
}

func TestSelectorString(t *testing.T) {
	tests := []struct {
		name     string
		selector document.Selector
		expected string
	}{
		{
			name:     "unconditional",
			selector: document.Selector{},
			expected: "No selection conditions specified",
		},
		{
			name:     "by-name",
			selector: document.NewSelector().ByName("foo"),
			expected: `[Name="foo"]`,
		},
		{
			name: "by-all",
			selector: document.NewSelector().
				ByGvk("testGroup", "testVersion", "testKind").
				ByNamespace("testNamespace").
				ByName("testName").
				ByAnnotation("testAnnotation=true").
				ByLabel("testLabel=true"),
			expected: `[Group="testGroup", Version="testVersion", Kind="testKind", ` +
				`Namespace="testNamespace", Name="testName", ` +
				`Annotations="testAnnotation=true", Labels="testLabel=true"]`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.selector.String())
		})
	}
}

func TestSelectorToObject(t *testing.T) {
	tests := []struct {
		name        string
		obj         runtime.Object
		expectedSel document.Selector
		expectedErr string
	}{
		{
			name: "Selector with GVK",
			obj:  &airapiv1.Clusterctl{},
			expectedSel: document.Selector{
				Selector: types.Selector{
					Gvk: resid.Gvk{
						Group:   "airshipit.org",
						Version: "v1alpha1",
						Kind:    "Clusterctl",
					},
				},
			},
			expectedErr: "",
		},
		{
			name:        "Unregistered object",
			obj:         &k8sv1.Pod{},
			expectedSel: document.Selector{},
			expectedErr: "no kind is registered for the type v1.Pod in scheme",
		},
		{
			name: "Selector with GVK and Name",
			obj: &airapiv1.Clusterctl{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterctl-v1",
				},
			},
			expectedSel: document.Selector{
				Selector: types.Selector{
					Gvk: resid.Gvk{
						Group:   "airshipit.org",
						Version: "v1alpha1",
						Kind:    "Clusterctl",
					},
					Name: "clusterctl-v1",
				},
			},
			expectedErr: "",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			actualSel, err := document.NewSelector().
				ByObject(tt.obj, airapiv1.Scheme)
			if test.expectedErr != "" {
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSel, actualSel)
			}
		})
	}
}

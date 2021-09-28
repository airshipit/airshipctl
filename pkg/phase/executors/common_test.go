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

package executors_test

import (
	"testing"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

func TestRegisterExecutor(t *testing.T) {
	testCases := []struct {
		name         string
		executorName string
		registry     map[schema.GroupVersionKind]ifc.ExecutorFactory
		expectedGVK  schema.GroupVersionKind
		expectedErr  error
	}{
		{
			name:         "register clusterctl executor",
			executorName: executors.Clusterctl,
			registry:     make(map[schema.GroupVersionKind]ifc.ExecutorFactory),
			expectedGVK: schema.GroupVersionKind{
				Group:   "airshipit.org",
				Version: "v1alpha1",
				Kind:    "Clusterctl",
			},
		},
		{
			name:         "register container executor",
			executorName: executors.GenericContainer,
			registry:     make(map[schema.GroupVersionKind]ifc.ExecutorFactory),
			expectedGVK: schema.GroupVersionKind{
				Group:   "airshipit.org",
				Version: "v1alpha1",
				Kind:    "GenericContainer",
			},
		},
		{
			name:         "register k8s applier executor",
			executorName: executors.KubernetesApply,
			registry:     make(map[schema.GroupVersionKind]ifc.ExecutorFactory),
			expectedGVK: schema.GroupVersionKind{
				Group:   "airshipit.org",
				Version: "v1alpha1",
				Kind:    "KubernetesApply",
			},
		},
		{
			name:         "register ephemeral executor",
			executorName: executors.Ephemeral,
			registry:     make(map[schema.GroupVersionKind]ifc.ExecutorFactory),
			expectedGVK: schema.GroupVersionKind{
				Group:   "airshipit.org",
				Version: "v1alpha1",
				Kind:    "BootConfiguration",
			},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			err := executors.RegisterExecutor(tt.executorName, tt.registry)
			require.NoError(t, err)

			_, found := tt.registry[tt.expectedGVK]
			assert.True(t, found)
		})
	}
}

// executorDoc converts string to document object
func executorDoc(t *testing.T, s string) document.Document {
	doc, err := document.NewDocumentFromBytes([]byte(s))
	require.NoError(t, err)
	require.NotNil(t, doc)
	return doc
}

// executorBundle converts string to bundle object
func executorBundle(t *testing.T, s string) document.Bundle {
	b, err := document.NewBundleFromBytes([]byte(s))
	require.NoError(t, err)
	require.NotNil(t, b)
	return b
}

func wrapError(err error) events.Event {
	return events.NewEvent().WithErrorEvent(events.ErrorEvent{
		Error: err,
	})
}

func testClusterMap(t *testing.T) clustermap.ClusterMap {
	doc, err := document.NewDocumentFromBytes([]byte(singleExecutorClusterMap))
	require.NoError(t, err)
	require.NotNil(t, doc)
	apiObj := v1alpha1.DefaultClusterMap()
	err = doc.ToAPIObject(apiObj, v1alpha1.Scheme)
	require.NoError(t, err)
	return clustermap.NewClusterMap(apiObj)
}

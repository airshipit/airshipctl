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

package clustermap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
)

func TestClusterMap(t *testing.T) {
	targetCluster := "target"
	ephemeraCluster := "ephemeral"
	workloadCluster := "workload"
	workloadClusterNoParent := "workload without parent"
	workloadClusterAPIRefName := "workload-cluster-api"
	workloadClusterAPIRefNamespace := "some-namespace"
	apiMap := &v1alpha1.ClusterMap{
		Map: map[string]*v1alpha1.Cluster{
			targetCluster: {
				Parent: ephemeraCluster,
				Sources: []v1alpha1.KubeconfigSource{
					{
						Type: v1alpha1.KubeconfigSourceTypeBundle,
					},
				},
			},
			ephemeraCluster: {},
			workloadCluster: {
				Parent: targetCluster,
				Sources: []v1alpha1.KubeconfigSource{
					{
						Type: v1alpha1.KubeconfigSourceTypeClusterAPI,
						ClusterAPI: v1alpha1.KubeconfigSourceClusterAPI{
							NamespacedName: v1alpha1.NamespacedName{
								Name:      workloadClusterAPIRefName,
								Namespace: workloadClusterAPIRefNamespace,
							},
						},
					},
				},
			},
			workloadClusterNoParent: {
				Sources: []v1alpha1.KubeconfigSource{
					{
						Type: v1alpha1.KubeconfigSourceTypeClusterAPI,
					},
				},
			},
		},
	}

	cMap := clustermap.NewClusterMap(apiMap)
	require.NotNil(t, cMap)

	t.Run("ephemeral parent", func(t *testing.T) {
		parent, err := cMap.ParentCluster(targetCluster)
		assert.NoError(t, err)
		assert.Equal(t, ephemeraCluster, parent)
	})

	t.Run("no cluster found", func(t *testing.T) {
		parent, err := cMap.ParentCluster("does not exist")
		assert.Error(t, err)
		assert.Equal(t, "", parent)
	})

	t.Run("target parent", func(t *testing.T) {
		parent, err := cMap.ParentCluster(workloadCluster)
		assert.NoError(t, err)
		assert.Equal(t, targetCluster, parent)
	})

	t.Run("ephemeral no parent", func(t *testing.T) {
		parent, err := cMap.ParentCluster(ephemeraCluster)
		assert.Error(t, err)
		assert.Equal(t, "", parent)
	})

	t.Run("all clusters", func(t *testing.T) {
		clusters := cMap.AllClusters()
		assert.Len(t, clusters, 4)
	})

	t.Run("kubeconfig context", func(t *testing.T) {
		kubeContext, err := cMap.ClusterKubeconfigContext(targetCluster)
		assert.NoError(t, err)
		assert.Equal(t, targetCluster, kubeContext)
	})

	t.Run("kubeconfig context error", func(t *testing.T) {
		_, err := cMap.ClusterKubeconfigContext("does not exist")
		assert.Error(t, err)
	})

	t.Run("sources match", func(t *testing.T) {
		sources, err := cMap.Sources(workloadCluster)
		assert.NoError(t, err)
		expectedSources := apiMap.Map[workloadCluster].Sources
		assert.Equal(t, expectedSources, sources)
	})

	t.Run("sources no cluster found", func(t *testing.T) {
		_, err := cMap.Sources("does not exist")
		assert.Error(t, err)
	})
}

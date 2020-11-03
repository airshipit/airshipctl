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
	workloadClusterKubeconfigContext := "different-workload-context"
	workloadClusterNoParent := "workload without parent"
	apiMap := &v1alpha1.ClusterMap{
		Map: map[string]*v1alpha1.Cluster{
			targetCluster: {
				Parent:            ephemeraCluster,
				DynamicKubeConfig: false,
			},
			ephemeraCluster: {},
			workloadCluster: {
				Parent:            targetCluster,
				DynamicKubeConfig: true,
				KubeconfigContext: workloadClusterKubeconfigContext,
			},
			workloadClusterNoParent: {
				DynamicKubeConfig: true,
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

	t.Run("not dynamic kubeconf target", func(t *testing.T) {
		dynamic := cMap.DynamicKubeConfig(targetCluster)
		assert.False(t, dynamic)
	})

	t.Run("dynamic kubeconf workload", func(t *testing.T) {
		dynamic := cMap.DynamicKubeConfig(workloadCluster)
		assert.True(t, dynamic)
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
		kubeContext, err := cMap.ClusterKubeconfigContext(workloadCluster)
		assert.NoError(t, err)
		assert.Equal(t, workloadClusterKubeconfigContext, kubeContext)
	})

	t.Run("kubeconfig default context", func(t *testing.T) {
		kubeContext, err := cMap.ClusterKubeconfigContext(targetCluster)
		assert.NoError(t, err)
		assert.Equal(t, targetCluster, kubeContext)
	})

	t.Run("kubeconfig context error", func(t *testing.T) {
		_, err := cMap.ClusterKubeconfigContext("does not exist")
		assert.Error(t, err)
	})
}

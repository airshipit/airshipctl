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

package kubeconfig_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/util"
)

func TestBuilder(t *testing.T) {
	t.Run("Only bundle", func(t *testing.T) {
		builder := kubeconfig.NewBuilder().WithBundle("testdata")
		kube := builder.Build()
		require.NotNil(t, kube)
		buf := bytes.NewBuffer([]byte{})
		err := kube.Write(buf)
		require.NoError(t, err)
		// check that kubeconfig contains expected cluster string
		assert.Contains(t, buf.String(), "dummycluster_ephemeral")
	})

	t.Run("Only filepath", func(t *testing.T) {
		builder := kubeconfig.NewBuilder().WithPath("testdata/kubeconfig")
		kube := builder.Build()
		require.NotNil(t, kube)
		buf := bytes.NewBuffer([]byte{})
		err := kube.Write(buf)
		require.NoError(t, err)
		// check that kubeconfig contains expected cluster string
		assert.Contains(t, buf.String(), "dummycluster_ephemeral")
	})

	t.Run("Only cluster map", func(t *testing.T) {
		childCluster := "child"
		parentCluster := "parent"
		clusterMap := &v1alpha1.ClusterMap{
			Map: map[string]*v1alpha1.Cluster{
				childCluster: {
					Parent:            parentCluster,
					DynamicKubeConfig: true,
				},
				parentCluster: {
					DynamicKubeConfig: false,
				},
			},
		}
		builder := kubeconfig.NewBuilder().
			WithClusterMap(clusterMap).
			WithClusterName(childCluster)
		kube := builder.Build()
		// This should not be implemented yet, and we need to check that we are getting there
		require.NotNil(t, kube)
		filePath, cleanup, err := kube.GetFile()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not implemented")
		assert.Equal(t, "", filePath)
		require.Nil(t, cleanup)
	})

	t.Run("No current cluster, fall to default", func(t *testing.T) {
		clusterMap := &v1alpha1.ClusterMap{}
		builder := kubeconfig.NewBuilder().
			WithClusterMap(clusterMap).
			WithClusterName("some-cluster")
		kube := builder.Build()
		// We should get a default value for cluster since we don't have some-cluster set
		actualPath, cleanup, err := kube.GetFile()
		require.NoError(t, err)
		defer cleanup()
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, kubeconfig.KubeconfigDefaultFileName)
		assert.Equal(t, path, actualPath)
	})

	t.Run("No parent cluster is defined, fall to default", func(t *testing.T) {
		childCluster := "child"
		clusterMap := &v1alpha1.ClusterMap{
			Map: map[string]*v1alpha1.Cluster{
				childCluster: {
					DynamicKubeConfig: true,
				},
			},
		}
		builder := kubeconfig.NewBuilder().
			WithClusterMap(clusterMap).
			WithClusterName(childCluster)
		kube := builder.Build()
		// We should get a default value for cluster, as we can't find parent cluster
		actualPath, cleanup, err := kube.GetFile()
		defer cleanup()
		require.NoError(t, err)
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, kubeconfig.KubeconfigDefaultFileName)
		assert.Equal(t, path, actualPath)
	})

	t.Run("Default source", func(t *testing.T) {
		builder := kubeconfig.NewBuilder()
		kube := builder.Build()
		// When ClusterMap is specified, but it doesn't have cluster-name defined, and no
		// other sources provided,
		actualPath, cleanup, err := kube.GetFile()
		defer cleanup()
		require.NoError(t, err)
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, kubeconfig.KubeconfigDefaultFileName)
		assert.Equal(t, path, actualPath)
	})
}

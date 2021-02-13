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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/testutil/clusterctl"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
)

const (
	testKubeconfigString = `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: c29tZWNlcnQK
    server: https://10.23.25.101:6443
  name: child_cluster
contexts:
- context:
    cluster: child_cluster
    user: child_user
  name: child
current-context: dummy_cluster
preferences: {}
users:
- name: child_user
  user:
    client-certificate-data: c29tZWNlcnQK
    client-key-data: c29tZWNlcnQK`
	testKubeconfigStringSecond = `apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: c29tZWNlcnQK
    server: https://10.23.25.101:6443
  name: parent_cluster
contexts:
- context:
    cluster: parent_cluster
    user: parent_admin
  name: parent-context
current-context: dummy_cluster
preferences: {}
users:
- name: parent_admin
  user:
    client-certificate-data: c29tZWNlcnQK
    client-key-data: c29tZWNlcnQK`
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
		assert.Contains(t, buf.String(), "parent_parent_context")
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

	t.Run("No current cluster, fall to default", func(t *testing.T) {
		clusterMap := &v1alpha1.ClusterMap{}
		builder := kubeconfig.NewBuilder().
			WithClusterMap(clustermap.NewClusterMap(clusterMap)).
			WithClusterName("some-cluster")
		kube := builder.Build()
		// We should get a default value for cluster since we don't have some-cluster set
		actualPath, cleanup, err := kube.GetFile()
		require.NoError(t, err)
		defer cleanup()
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, kubeconfig.KubeconfigDefaultFileName)
		assert.Equal(t, path, actualPath)
	})

	t.Run("Default source", func(t *testing.T) {
		builder := kubeconfig.NewBuilder()
		kube := builder.Build()
		// When ClusterMap is specified, but it doesn't have cluster-name defined, and no
		// other sources provided,
		actualPath, cleanup, err := kube.GetFile()
		require.NoError(t, err)
		defer cleanup()
		path := filepath.Join(util.UserHomeDir(), config.AirshipConfigDir, kubeconfig.KubeconfigDefaultFileName)
		assert.Equal(t, path, actualPath)
	})
}

func TestBuilderClusterctl(t *testing.T) {
	childCluster := "child"
	parentCluster := "parent"
	testBundlePath := "testdata"
	kubeconfigPath := filepath.Join(testBundlePath, "kubeconfig-12341234")

	tests := []struct {
		name                 string
		errString            string
		requestedClusterName string
		parentClusterName    string
		tempRoot             string
		clusterMap           clustermap.ClusterMap
		clusterctlClient     client.Interface
		fs                   fs.FileSystem
	}{
		{
			name:                 "error no parent context",
			errString:            fmt.Sprintf("context \"%s\" does not exist", parentCluster),
			parentClusterName:    parentCluster,
			requestedClusterName: childCluster,
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childCluster: {
						Parent:            parentCluster,
						DynamicKubeConfig: true,
					},
					parentCluster: {
						DynamicKubeConfig: false,
					},
				},
			}),
		},
		{
			name:                 "error dynamic but no parrent",
			parentClusterName:    parentCluster,
			requestedClusterName: childCluster,
			errString:            "failed to find a parent",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childCluster: {
						DynamicKubeConfig: true,
					},
				},
			}),
		},
		{
			name:                 "error write temp parent",
			parentClusterName:    parentCluster,
			requestedClusterName: childCluster,
			tempRoot:             "does not exist anywhere",
			errString:            "no such file or directory",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childCluster: {
						Parent:            parentCluster,
						DynamicKubeConfig: true,
					},
					parentCluster: {
						DynamicKubeConfig: false,
						KubeconfigContext: "dummy_cluster",
					},
				},
			}),
		},
		{
			name:                 "success",
			parentClusterName:    parentCluster,
			requestedClusterName: childCluster,
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childCluster: {
						Parent:            parentCluster,
						DynamicKubeConfig: true,
					},
					parentCluster: {
						DynamicKubeConfig: true,
						KubeconfigContext: "parent-custom",
						Parent:            "parent-parent",
					},
					"parent-parent": {
						DynamicKubeConfig: false,
						KubeconfigContext: "parent_parent_context",
					},
				},
			}),
			tempRoot: "testdata",
			fs: testfs.MockFileSystem{
				MockRemoveAll: func() error { return nil },
				MockTempFile: func(s1, s2 string) (fs.File, error) {
					return testfs.TestFile{
						MockName:  func() string { return kubeconfigPath },
						MockWrite: func() (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
			},
			clusterctlClient: func() client.Interface {
				c := &clusterctl.MockInterface{
					Mock: mock.Mock{},
				}
				c.On("GetKubeconfig", &client.GetKubeconfigOptions{
					ParentKubeconfigPath:    kubeconfigPath,
					ParentKubeconfigContext: "parent-custom",
					ManagedClusterNamespace: clustermap.DefaultClusterAPIObjNamespace,
					ManagedClusterName:      childCluster,
				}).Once().Return(testKubeconfigString, nil)
				c.On("GetKubeconfig", &client.GetKubeconfigOptions{
					ParentKubeconfigPath:    kubeconfigPath,
					ParentKubeconfigContext: "parent_parent_context",
					ManagedClusterNamespace: clustermap.DefaultClusterAPIObjNamespace,
					ManagedClusterName:      parentCluster,
				}).Once().Return(testKubeconfigStringSecond, nil)
				return c
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kube := kubeconfig.NewBuilder().
				WithClusterMap(tt.clusterMap).
				WithClusterName(tt.requestedClusterName).
				WithBundle(testBundlePath).
				WithTempRoot(tt.tempRoot).
				WithClusterctClient(tt.clusterctlClient).
				WithFilesytem(tt.fs).
				Build()
			require.NotNil(t, kube)
			filePath, cleanup, err := kube.GetFile()
			// This is needed to avoid leftovers on test failures
			if cleanup != nil {
				defer cleanup()
			}

			if tt.errString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				assert.Equal(t, "", filePath)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, "", filePath)
				assert.NotNil(t, cleanup)
				buf := bytes.NewBuffer([]byte{})
				err := kube.Write(buf)
				require.NoError(t, err)
				result, err := ioutil.ReadFile("testdata/result-kubeconf")
				require.NoError(t, err)
				assert.Equal(t, result, buf.Bytes())
			}
		})
	}
}

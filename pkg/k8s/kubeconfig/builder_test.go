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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
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

func TestBuilderClusterctl(t *testing.T) {
	childClusterID := "child"
	parentClusterID := "parent"
	parentParentClusterID := "parent-parent"
	// these are values in kubeconfig.cluster
	parentCluster := "parent_cluster"
	parentParentCluster := "parent_parent_cluster"
	childCluster := "child_cluster"
	parentUser := "parent_admin"
	parentParentUser := "parent_parent_admin"
	childUser := "child_user"
	testBundle, err := document.NewBundleByPath("testdata")
	require.NoError(t, err)
	kubeconfigPath := filepath.Join("testdata", "kubeconfig-12341234")

	tests := []struct {
		name                 string
		errString            string
		requestedClusterName string
		tempRoot             string
		siteWide             bool

		expectedContexts, expectedClusters, expectedAuthInfos []string
		clusterMap                                            clustermap.ClusterMap
		clusterctlClient                                      client.Interface
		fs                                                    fs.FileSystem
	}{
		{
			name:              "success cluster-api not reachable",
			expectedContexts:  []string{parentClusterID},
			expectedClusters:  []string{parentParentCluster},
			expectedAuthInfos: []string{parentParentUser},
			siteWide:          true,
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childClusterID: {
						Parent: parentClusterID,
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeClusterAPI,
							},
						},
					},
					parentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeBundle,
								Bundle: v1alpha1.KubeconfigSourceBundle{
									Context: "parent_parent_context",
								},
							},
						},
					},
				},
			}),
		},
		{
			name:              "success two clusters",
			expectedContexts:  []string{parentClusterID, parentParentClusterID},
			expectedClusters:  []string{"dummycluster_ephemeral", parentParentCluster},
			expectedAuthInfos: []string{"kubernetes-admin", parentParentUser},
			siteWide:          true,
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					parentParentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeBundle,
								Bundle: v1alpha1.KubeconfigSourceBundle{
									Context: "parent_parent_context",
								},
							},
						},
					},
					parentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeFilesystem,
								FileSystem: v1alpha1.KubeconfigSourceFilesystem{
									Path:    "testdata/kubeconfig",
									Context: "dummy_cluster",
								},
							},
						},
					},
				},
			}),
		},
		{
			name:              "success three clusters cluster-api",
			expectedContexts:  []string{parentClusterID, childClusterID, parentParentClusterID},
			expectedClusters:  []string{parentCluster, parentParentCluster, childCluster},
			expectedAuthInfos: []string{parentUser, parentParentUser, childUser},
			siteWide:          true,
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					childClusterID: {
						Parent: parentClusterID,
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeClusterAPI,
								ClusterAPI: v1alpha1.KubeconfigSourceClusterAPI{
									NamespacedName: v1alpha1.NamespacedName{
										Name:      childClusterID,
										Namespace: "target-infra",
									},
								},
							},
						},
					},
					parentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeClusterAPI,
								ClusterAPI: v1alpha1.KubeconfigSourceClusterAPI{
									NamespacedName: v1alpha1.NamespacedName{
										Name:      parentClusterID,
										Namespace: "target-infra",
									},
								},
							},
						},
						Parent: parentParentClusterID,
					},
					parentParentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeBundle,
							},
						},
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
					ParentKubeconfigContext: parentClusterID,
					ManagedClusterNamespace: clustermap.DefaultClusterAPIObjNamespace,
					ManagedClusterName:      childClusterID,
				}).Once().Return(testKubeconfigString, nil)
				c.On("GetKubeconfig", &client.GetKubeconfigOptions{
					ParentKubeconfigPath:    kubeconfigPath,
					ParentKubeconfigContext: parentParentClusterID,
					ManagedClusterNamespace: clustermap.DefaultClusterAPIObjNamespace,
					ManagedClusterName:      parentClusterID,
				}).Once().Return(testKubeconfigStringSecond, nil)
				return c
			}(),
		},
		{
			name:                 "error requested cluster doesn't exist",
			errString:            "is not defined in cluster map",
			requestedClusterName: "non-existent-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					parentParentClusterID: {
						Sources: []v1alpha1.KubeconfigSource{
							{
								Type: v1alpha1.KubeconfigSourceTypeBundle,
								Bundle: v1alpha1.KubeconfigSourceBundle{
									Context: "parent_parent_context",
								},
							},
						},
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			kube := kubeconfig.NewBuilder().
				WithClusterMap(tt.clusterMap).
				WithClusterName(tt.requestedClusterName).
				WithBundle(testBundle).
				WithTempRoot(tt.tempRoot).
				WithClusterctlClient(tt.clusterctlClient).
				WithFilesystem(tt.fs).
				SiteWide(tt.siteWide).
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
				require.NoError(t, err)
				assert.NotEqual(t, "", filePath)
				assert.NotNil(t, cleanup)
				buf := bytes.NewBuffer([]byte{})
				err := kube.Write(buf)
				require.NoError(t, err)
				compareResults(t, tt.expectedContexts, tt.expectedClusters, tt.expectedAuthInfos, buf.Bytes())
			}
		})
	}
}

func compareResults(t *testing.T, contexts, clusters, authInfos []string, kubeconfBytes []byte) {
	t.Helper()
	resultKubeconf, err := clientcmd.Load(kubeconfBytes)
	require.NoError(t, err)

	assert.Len(t, resultKubeconf.Contexts, len(contexts))
	for _, name := range contexts {
		assert.Contains(t, resultKubeconf.Contexts, name)
	}

	assert.Len(t, resultKubeconf.AuthInfos, len(authInfos))
	for _, name := range authInfos {
		assert.Contains(t, resultKubeconf.AuthInfos, name)
	}

	assert.Len(t, resultKubeconf.Clusters, len(clusters))
	for _, name := range clusters {
		assert.Contains(t, resultKubeconf.Clusters, name)
	}
}

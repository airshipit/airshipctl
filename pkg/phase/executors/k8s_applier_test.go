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
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
)

const (
	ValidExecutorDoc = `apiVersion: airshipit.org/v1alpha1
kind: KubernetesApply
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: kubernetes-apply
config:
  waitOptions:
    timeout: 600
  pruneOptions:
    prune: false
`
	WrongExecutorDoc = `apiVersion: v1
kind: ConfigMap
metadata:
  name: first-map
  namespace: default
  labels:
    cli-utils.sigs.k8s.io/inventory-id: "some id"
`
	applierKRMDoc = `---
apiVersion: airshipit.org/v1alpha1
kind: GenericContainer
metadata:
  name: applier
  labels:
    airshipit.org/deploy-k8s: "false"
spec:
  type: krm
  image: quay.io/airshipit/applier:latest
  hostNetwork: true
`
)

func testApplierBundleFactory(t *testing.T, filteredContent string, writer io.Writer) document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		// When the k8s applier executor Run method is called, the executor bundle is filtered
		// using the label selector "airshipit.org/deploy-k8s notin (False, false)".
		// Render method just filters out document with a given selector.
		// That is why we need "SelectBundle" method mocked and return a filtered bundle.
		filteredBundle := &testdoc.MockBundle{}
		// Filtered bundle is passed to the k8s applier, which looks it up
		// for an inventory document using label "cli-utils.sigs.k8s.io/inventory-id"
		// and kind "ConfigMap". Therefore "SelectOne" method is mocked in the filtered bundle.
		// This mock is used to get inventory document. Empty document is ok for k8s applier.
		filteredBundle.On("SelectOne", mock.Anything).Return(&testdoc.MockDocument{}, nil)
		// This mock is used to get the contents of the applier filtered bundle both for
		// rendering and for applying it to a k8s cluster.
		filteredBundle.On("Write", writer).
			Return(nil).
			Run(func(args mock.Arguments) {
				arg, ok := args.Get(0).(io.Writer)
				if ok {
					_, err := arg.Write([]byte(filteredContent))
					require.NoError(t, err)
				}
			})
		// This is the applier executor bundle.
		bundle := &testdoc.MockBundle{}
		// This mock is used to filter out documents labeled with
		// "airshipit.org/deploy-k8s notin (False, false)"
		bundle.On("SelectBundle", mock.Anything).Return(filteredBundle, nil)
		return bundle, nil
	}
}

func testApplierBundleFactorySelectBundleError() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		bundle.On("SelectOne", mock.Anything).
			Return(nil, nil)
		bundle.On("SelectBundle", mock.Anything).
			Return(nil, errors.New("error selecting bundle"))
		return bundle, nil
	}
}

func testApplierBundleFactoryWriteError() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		filteredBundle := &testdoc.MockBundle{}
		bundle.On("SelectOne", mock.Anything).
			Return(nil, nil)
		filteredBundle.On("Write", mock.Anything).
			Return(errors.New("error writing bundle"))
		bundle.On("SelectBundle", mock.Anything).
			Return(filteredBundle, nil)
		return bundle, nil
	}
}

func testApplierBundleFactoryNoError() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		filteredBundle := &testdoc.MockBundle{}
		bundle.On("SelectOne", mock.Anything).
			Return(nil, nil)
		filteredBundle.On("Write", mock.Anything).
			Return(nil)
		bundle.On("SelectBundle", mock.Anything).
			Return(filteredBundle, nil)
		return bundle, nil
	}
}

func testApplierBundleFactoryEmptyAllDocuments() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		bundle.On("GetAllDocuments").Return([]document.Document{}, nil)
		return bundle, nil
	}
}

func testApplierBundleFactoryAllDocuments() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		bundle.On("GetAllDocuments").Return([]document.Document{&testdoc.MockDocument{}}, nil)
		return bundle, nil
	}
}

func TestNewKubeApplierExecutor(t *testing.T) {
	tests := []struct {
		name          string
		execDoc       document.Document
		expectedErr   bool
		bundleFactory document.BundleFactoryFunc
		phaseBundle   document.Bundle
	}{
		{
			name:        "invalid executor document",
			execDoc:     executorDoc(t, WrongExecutorDoc),
			expectedErr: true,
		},
		{
			name:          "invalid bundle factory",
			execDoc:       executorDoc(t, ValidExecutorDoc),
			bundleFactory: testdoc.ErrorBundleFactory,
			expectedErr:   true,
		},
		{
			name:          "invalid phase config bundle",
			execDoc:       executorDoc(t, ValidExecutorDoc),
			bundleFactory: testdoc.EmptyBundleFactory,
			phaseBundle:   executorBundle(t, ""),
			expectedErr:   true,
		},
		{
			name:          "valid phase config bundle",
			execDoc:       executorDoc(t, ValidExecutorDoc),
			bundleFactory: testdoc.EmptyBundleFactory,
			phaseBundle:   executorBundle(t, applierKRMDoc),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			exec, err := executors.NewKubeApplierExecutor(
				ifc.ExecutorConfig{
					ExecutorDocument:  tt.execDoc,
					BundleFactory:     tt.bundleFactory,
					PhaseConfigBundle: tt.phaseBundle,
				})
			if tt.expectedErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "")
				assert.Nil(t, exec)
			} else {
				require.NoError(t, err)
				require.NotNil(t, exec)
			}
		})
	}
}

// TODO We need valid test that checks that actual bundle has arrived to applier
// for that we need a way to inject fake applier, which is not doable with `black box` test currently
// since we tests are in different package from executor
func TestKubeApplierExecutorRun(t *testing.T) {
	tests := []struct {
		name        string
		containsErr string
		clusterName string

		kubeconf      kubeconfig.Interface
		execDoc       document.Document
		bundleFactory document.BundleFactoryFunc
		clusterMap    clustermap.ClusterMap
		clientFunc    container.ClientV1Alpha1FactoryFunc
	}{
		{
			name:          "unable to get kubeconfig context",
			containsErr:   "cluster 'foo' is not defined in cluster map",
			bundleFactory: testApplierBundleFactory(t, "", nil),
			kubeconf:      testKubeconfig(`invalid kubeconfig`),
			execDoc:       executorDoc(t, ValidExecutorDoc),
			clusterName:   "foo",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
		},
		{
			name:          "unable to get kubeconfig file",
			containsErr:   "failed to get kubeconfig",
			bundleFactory: testApplierBundleFactory(t, "", nil),
			kubeconf: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", nil, errors.New("failed to get kubeconfig")
			}},
			execDoc:     executorDoc(t, ValidExecutorDoc),
			clusterName: "ephemeral-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
		},
		{
			name:          "unable to select bundle",
			containsErr:   "error selecting bundle",
			bundleFactory: testApplierBundleFactorySelectBundleError(),
			kubeconf:      testKubeconfig("kubeconfig"),
			execDoc:       executorDoc(t, ValidExecutorDoc),
			clusterName:   "ephemeral-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
		},
		{
			name:          "unable to write bundle",
			containsErr:   "error writing bundle",
			bundleFactory: testApplierBundleFactoryWriteError(),
			kubeconf:      testKubeconfig("kubeconfig"),
			execDoc:       executorDoc(t, ValidExecutorDoc),
			clusterName:   "ephemeral-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
		},
		{
			name:          "unsuccessful run",
			containsErr:   "applier failure",
			bundleFactory: testApplierBundleFactoryNoError(),
			kubeconf:      testKubeconfig("kubeconfig"),
			execDoc:       executorDoc(t, ValidExecutorDoc),
			clusterName:   "ephemeral-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
			clientFunc: func(_ string, _ io.Reader, _ io.Writer,
				_ *v1alpha1.GenericContainer, _ string) container.ClientV1Alpha1 {
				return MockClientFuncInterface{MockRun: func() error {
					return errors.New("applier failure")
				}}
			},
		},
		{
			name:          "successful run",
			containsErr:   "",
			bundleFactory: testApplierBundleFactoryNoError(),
			kubeconf:      testKubeconfig("kubeconfig"),
			execDoc:       executorDoc(t, ValidExecutorDoc),
			clusterName:   "ephemeral-cluster",
			clusterMap: clustermap.NewClusterMap(&v1alpha1.ClusterMap{
				Map: map[string]*v1alpha1.Cluster{
					"ephemeral-cluster": {},
				},
			}),
			clientFunc: func(_ string, _ io.Reader, _ io.Writer,
				_ *v1alpha1.GenericContainer, _ string) container.ClientV1Alpha1 {
				return MockClientFuncInterface{MockRun: func() error {
					return nil
				}}
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			exec, err := executors.NewKubeApplierExecutor(
				ifc.ExecutorConfig{
					ExecutorDocument:  tt.execDoc,
					BundleFactory:     tt.bundleFactory,
					KubeConfig:        tt.kubeconf,
					ClusterMap:        tt.clusterMap,
					ClusterName:       tt.clusterName,
					PhaseConfigBundle: executorBundle(t, applierKRMDoc),
					ContainerFunc:     tt.clientFunc,
				})
			require.NoError(t, err)
			require.NotNil(t, exec)
			ch := make(chan events.Event)
			go exec.Run(ch, ifc.RunOptions{})
			processor := events.NewDefaultProcessor()
			err = processor.Process(ch)
			if tt.containsErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.containsErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRender(t *testing.T) {
	writer := bytes.NewBuffer([]byte{})
	content := "Some content"
	exec, err := executors.NewKubeApplierExecutor(ifc.ExecutorConfig{
		BundleFactory:     testApplierBundleFactory(t, content, writer),
		ExecutorDocument:  executorDoc(t, ValidExecutorDoc),
		PhaseConfigBundle: executorBundle(t, applierKRMDoc),
	})
	require.NoError(t, err)
	require.NotNil(t, exec)

	err = exec.Render(writer, ifc.RenderOptions{})
	require.NoError(t, err)

	result := writer.String()
	assert.Equal(t, content, result)
}

func testKubeconfig(stringData string) kubeconfig.Interface {
	return kubeconfig.NewKubeConfig(
		kubeconfig.FromByte([]byte(stringData)),
		kubeconfig.InjectFileSystem(
			testfs.MockFileSystem{
				MockTempFile: func(root, pattern string) (fs.File, error) {
					return testfs.TestFile{
						MockName:  func() string { return "kubeconfig-142398" },
						MockWrite: func([]byte) (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
				MockRemoveAll: func() error { return nil },
			},
		))
}

func TestKubeApplierExecutor_Validate(t *testing.T) {
	tests := []struct {
		name          string
		bundleFactory document.BundleFactoryFunc
		bundleName    string
		wantErr       bool
	}{
		{
			name:          "Error empty BundleName",
			bundleFactory: testdoc.EmptyBundleFactory,
			wantErr:       true,
		},
		{
			name:          "Error no documents",
			bundleName:    "some name",
			bundleFactory: testApplierBundleFactoryEmptyAllDocuments(),
			wantErr:       true,
		},
		{
			name:          "Success case",
			bundleName:    "some name",
			bundleFactory: testApplierBundleFactoryAllDocuments(),
			wantErr:       false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			e, err := executors.NewKubeApplierExecutor(ifc.ExecutorConfig{
				BundleFactory:     tt.bundleFactory,
				PhaseName:         tt.bundleName,
				ExecutorDocument:  executorDoc(t, ValidExecutorDoc),
				PhaseConfigBundle: executorBundle(t, applierKRMDoc),
			})
			require.NoError(t, err)
			require.NotNil(t, e)

			if err := e.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("KubeApplierExecutor.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

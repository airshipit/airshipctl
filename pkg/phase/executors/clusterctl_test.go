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
	goerrors "errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/executors/errors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

var (
	executorConfigTmpl = `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl-v1
action: %s
init-options:
  core-provider: "cluster-api:v0.3.2"
  kubeConfigRef:
    apiVersion: airshipit.org/v1alpha1
    kind: KubeConfig
    name: sample-name
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2`

	executorConfigTmplGood = `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl-v1
action: %s
init-options:
  core-provider: "cluster-api:v0.3.2"
  kubeConfigRef:
    apiVersion: airshipit.org/v1alpha1
    kind: KubeConfig
    name: sample-name
  namespace: some-namespace
move-options:
  namespace: some-namespace
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2`

	renderedDocs = `---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    control-plane: controller-manager
  name: version-two
...
`
	krmExecDoc = `---
apiVersion: airshipit.org/v1alpha1
kind: GenericContainer
metadata:
  name: clusterctl
  labels:
    airshipit.org/deploy-k8s: "false"
spec:
  type: krm
  image: localhost/clusterctl:latest
  hostNetwork: true
`
)

func TestNewClusterctlExecutor(t *testing.T) {
	testCases := []struct {
		name           string
		targetPath     string
		phaseCfgBundle document.Bundle
		execDoc        document.Document
		errContains    string
	}{
		{
			name: "invalid executor document",
			execDoc: executorDoc(t, `
apiVersion: test.org/v1alpha1
kind: testkind
metadata:
  name: testname
`),
			errContains: "no kind \"testkind\" is registered for version \"test.org/v1alpha1\"",
		},
		{
			name:       "broken kustomize entrypoint",
			targetPath: "/not/exist",
			execDoc: executorDoc(t, `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2"
`),
			errContains: "no such file or directory",
		},
		{
			name:       "no metadata available",
			targetPath: "./testdata",
			execDoc: executorDoc(t, `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2/no-metadata
`),
			errContains: "document filtered by selector " +
				"[Group=\"clusterctl.cluster.x-k8s.io\", Version=\"v1alpha3\", Kind=\"Metadata\"] found no documents",
		},
		{
			name:           "no container executor available",
			targetPath:     "./testdata",
			phaseCfgBundle: executorBundle(t, ""),
			execDoc: executorDoc(t, `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2
`),
			errContains: "document filtered by selector [Group=\"airshipit.org\", " +
				"Version=\"v1alpha1\", Kind=\"GenericContainer\", Name=\"clusterctl\"] found no documents",
		},
		{
			name:           "successfully create executor",
			targetPath:     "./testdata",
			phaseCfgBundle: executorBundle(t, krmExecDoc),
			execDoc: executorDoc(t, `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    url: functions/capi/v0.3.2
`),
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			executor, actualErr := executors.NewClusterctlExecutor(ifc.ExecutorConfig{
				ExecutorDocument:  tt.execDoc,
				TargetPath:        tt.targetPath,
				PhaseConfigBundle: tt.phaseCfgBundle,
			})
			if tt.errContains != "" {
				require.Nil(t, executor)
				require.NotNil(t, actualErr)
				require.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				require.NotNil(t, executor)
			}
		})
	}
}

var _ clustermap.ClusterMap = &ClusterMapMockInterface{}

type ClusterMapMockInterface struct {
	MockClusterKubeconfigContext func(string) (string, error)
	MockParentCluster            func(string) (string, error)
}

func (c ClusterMapMockInterface) ValidateClusterMap() error {
	panic("implement me")
}

func (c ClusterMapMockInterface) ParentCluster(s string) (string, error) {
	return c.MockParentCluster(s)
}

func (c ClusterMapMockInterface) AllClusters() []string {
	panic("implement me")
}

func (c ClusterMapMockInterface) ClusterKubeconfigContext(s string) (string, error) {
	return c.MockClusterKubeconfigContext(s)
}

func (c ClusterMapMockInterface) Sources(_ string) ([]v1alpha1.KubeconfigSource, error) {
	panic("implement me")
}

func (c ClusterMapMockInterface) Write(_ io.Writer, _ clustermap.WriteOptions) error {
	panic("implement me")
}

var _ container.ClientV1Alpha1 = &MockClientFuncInterface{}

type MockClientFuncInterface struct {
	MockRun func() error
}

func (c MockClientFuncInterface) Run() error {
	return c.MockRun()
}

func TestClusterctlExecutorRun(t *testing.T) {
	errTmpFile := goerrors.New("TmpFile error")
	errCtx := goerrors.New("context error")
	errParent := goerrors.New("parent cluster error")
	testCases := []struct {
		name        string
		cfgDoc      document.Document
		kubecfg     kubeconfig.Interface
		expectedErr error
		clusterMap  clustermap.ClusterMap
		clientFunc  container.ClientV1Alpha1FactoryFunc
	}{
		{
			name:        "Error unknown action",
			cfgDoc:      executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "someAction")),
			expectedErr: errors.ErrUnknownExecutorAction{Action: "someAction", ExecutorName: "clusterctl"},
			clusterMap:  clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
		},
		{
			name:   "Failed get kubeconfig file - init",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "init")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", nil, errTmpFile
			}},
			expectedErr: errTmpFile,
			clusterMap:  clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
		},
		{
			name:   "Failed get kubeconfig file - move",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "move")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", nil, errTmpFile
			}},
			expectedErr: errTmpFile,
			clusterMap:  clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
		},
		{
			name:   "Failed get kubeconfig context - init",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "init")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", func() {}, nil
			}},
			expectedErr: errCtx,
			clusterMap: ClusterMapMockInterface{MockClusterKubeconfigContext: func(s string) (string, error) {
				return "", errCtx
			}},
		},
		{
			name:   "Failed get kubeconfig context - move",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "move")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", func() {}, nil
			}},
			expectedErr: errCtx,
			clusterMap: ClusterMapMockInterface{MockClusterKubeconfigContext: func(s string) (string, error) {
				return "", errCtx
			}},
		},
		{
			name:   "Failed get parent cluster",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "move")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", func() {}, nil
			}},
			expectedErr: errParent,
			clusterMap: ClusterMapMockInterface{MockClusterKubeconfigContext: func(s string) (string, error) {
				return "ctx", nil
			},
				MockParentCluster: func(s string) (string, error) {
					return "", errParent
				}},
		},
		{
			name:   "Regular Run init",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "init")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", func() {}, nil
			}},
			clusterMap: ClusterMapMockInterface{MockClusterKubeconfigContext: func(s string) (string, error) {
				return "cluster", nil
			}},
			clientFunc: func(_ string, _ io.Reader, _ io.Writer,
				_ *v1alpha1.GenericContainer, _ string) container.ClientV1Alpha1 {
				return MockClientFuncInterface{MockRun: func() error {
					return nil
				}}
			},
		},
		{
			name:   "Regular Run move",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmplGood, "move")),
			kubecfg: fakeKubeConfig{getFile: func() (string, kubeconfig.Cleanup, error) {
				return "", func() {}, nil
			}},
			clusterMap: ClusterMapMockInterface{MockClusterKubeconfigContext: func(s string) (string, error) {
				return "cluster", nil
			},
				MockParentCluster: func(s string) (string, error) {
					return "parentCluster", nil
				}},
			clientFunc: func(_ string, _ io.Reader, _ io.Writer,
				_ *v1alpha1.GenericContainer, _ string) container.ClientV1Alpha1 {
				return MockClientFuncInterface{MockRun: func() error {
					return nil
				}}
			},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			executor, err := executors.NewClusterctlExecutor(
				ifc.ExecutorConfig{
					TargetPath:        "testdata",
					PhaseConfigBundle: executorBundle(t, krmExecDoc),
					ExecutorDocument:  tt.cfgDoc,
					KubeConfig:        tt.kubecfg,
					ClusterMap:        tt.clusterMap,
					ContainerFunc:     tt.clientFunc,
				})
			require.NoError(t, err)
			err = executor.Run(ifc.RunOptions{DryRun: true})
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestClusterctlExecutorValidate(t *testing.T) {
	testCases := []struct {
		name               string
		actionType         string
		bundlePath         string
		executorConfigTmpl string
		expectedErrString  string
	}{
		{
			name:               "Success init action",
			actionType:         "init",
			executorConfigTmpl: executorConfigTmplGood,
		},
		{
			name:               "Error any other action",
			actionType:         "any",
			executorConfigTmpl: executorConfigTmplGood,
			expectedErrString:  "unknown action type 'any'",
		},
		{
			name:               "Error empty action",
			actionType:         "",
			executorConfigTmpl: executorConfigTmplGood,
			expectedErrString:  "invalid phase: ClusterctlExecutor.Action is empty",
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			sampleCfgDoc := executorDoc(t, fmt.Sprintf(test.executorConfigTmpl, test.actionType))
			executor, err := executors.NewClusterctlExecutor(
				ifc.ExecutorConfig{
					TargetPath:        "testdata",
					ExecutorDocument:  sampleCfgDoc,
					PhaseConfigBundle: executorBundle(t, krmExecDoc),
				})
			require.NoError(t, err)
			err = executor.Validate()
			if test.expectedErrString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestClusterctlExecutorRender(t *testing.T) {
	sampleCfgDoc := executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init"))
	executor, err := executors.NewClusterctlExecutor(
		ifc.ExecutorConfig{
			TargetPath:        "testdata",
			ExecutorDocument:  sampleCfgDoc,
			PhaseConfigBundle: executorBundle(t, krmExecDoc),
		})
	require.NoError(t, err)
	actualOut := &bytes.Buffer{}
	actualErr := executor.Render(actualOut, ifc.RenderOptions{})
	assert.Equal(t, renderedDocs, actualOut.String())
	assert.NoError(t, actualErr)
}

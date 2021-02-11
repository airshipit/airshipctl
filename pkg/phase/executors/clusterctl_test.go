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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/cluster/clustermap"
	"opendev.org/airship/airshipctl/pkg/document"
	airerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/fs"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
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
    versions:
      v0.3.2: functions/capi/infrastructure/v0.3.2`

	renderedDocs = `---
apiVersion: v1
kind: Namespace
metadata:
  labels:
    cluster.x-k8s.io/provider: cluster-api
    clusterctl.cluster.x-k8s.io: ""
    control-plane: controller-manager
  name: version-two
...
`
)

func TestNewExecutor(t *testing.T) {
	sampleCfgDoc := executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init"))
	testCases := []struct {
		name        string
		helper      ifc.Helper
		expectedErr error
	}{
		{
			name:   "New Clusterctl Executor",
			helper: makeDefaultHelper(t, "../../clusterctl/client/testdata", defaultMetadataPath),
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			_, actualErr := executors.NewClusterctlExecutor(ifc.ExecutorConfig{
				ExecutorDocument: sampleCfgDoc,
				Helper:           tt.helper,
			})
			assert.Equal(t, tt.expectedErr, actualErr)
		})
	}
}

func TestExecutorRun(t *testing.T) {
	errTmpFile := errors.New("TmpFile error")

	testCases := []struct {
		name        string
		cfgDoc      document.Document
		fs          fs.FileSystem
		bundlePath  string
		expectedEvt []events.Event
		clusterMap  clustermap.ClusterMap
	}{
		{
			name:       "Error unknown action",
			cfgDoc:     executorDoc(t, fmt.Sprintf(executorConfigTmpl, "someAction")),
			bundlePath: "testdata/executor_init",
			expectedEvt: []events.Event{
				wrapError(executors.ErrUnknownExecutorAction{Action: "someAction", ExecutorName: "clusterctl"}),
			},
			clusterMap: clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
		},
		{
			name:   "Error temporary file",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init")),
			fs: testfs.MockFileSystem{
				MockTempFile: func(string, string) (fs.File, error) {
					return nil, errTmpFile
				},
			},
			bundlePath: "testdata/executor_init",
			expectedEvt: []events.Event{
				events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
					Operation: events.ClusterctlInitStart,
				}),
				wrapError(errTmpFile),
			},
			clusterMap: clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
		},
		{
			name:   "Regular Run init",
			cfgDoc: executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init")),
			fs: testfs.MockFileSystem{
				MockTempFile: func(string, string) (fs.File, error) {
					return testfs.TestFile{
						MockName:  func() string { return "filename" },
						MockWrite: func() (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
				MockRemoveAll: func() error { return nil },
			},
			bundlePath: "testdata/executor_init",
			clusterMap: clustermap.NewClusterMap(v1alpha1.DefaultClusterMap()),
			expectedEvt: []events.Event{
				events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
					Operation: events.ClusterctlInitStart,
				}),
				events.NewEvent().WithClusterctlEvent(events.ClusterctlEvent{
					Operation: events.ClusterctlInitEnd,
				}),
			},
		},
		// TODO add move tests here
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			kubeCfg := kubeconfig.NewKubeConfig(
				kubeconfig.FromByte([]byte("someKubeConfig")),
				kubeconfig.InjectFileSystem(tt.fs),
			)
			executor, err := executors.NewClusterctlExecutor(
				ifc.ExecutorConfig{
					ExecutorDocument: tt.cfgDoc,
					Helper:           makeDefaultHelper(t, "../../clusterctl/client/testdata", defaultMetadataPath),
					KubeConfig:       kubeCfg,
					ClusterMap:       tt.clusterMap,
				})
			require.NoError(t, err)
			ch := make(chan events.Event)
			go executor.Run(ch, ifc.RunOptions{DryRun: true})
			var actualEvt []events.Event
			for evt := range ch {
				// Skip timmestamp for comparison
				evt.Timestamp = time.Time{}
				if evt.Type == events.ClusterctlType {
					// Set message to empty string, so it's not compared
					evt.ClusterctlEvent.Message = ""
				}
				actualEvt = append(actualEvt, evt)
			}
			for i := range tt.expectedEvt {
				// Skip timmestamp for comparison
				tt.expectedEvt[i].Timestamp = time.Time{}
			}
			assert.Equal(t, tt.expectedEvt, actualEvt)
		})
	}
}

func TestExecutorValidate(t *testing.T) {
	sampleCfgDoc := executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init"))
	executor, err := executors.NewClusterctlExecutor(
		ifc.ExecutorConfig{
			ExecutorDocument: sampleCfgDoc,
			Helper:           makeDefaultHelper(t, "../../clusterctl/client/testdata", defaultMetadataPath),
		})
	require.NoError(t, err)
	expectedErr := airerrors.ErrNotImplemented{}
	actualErr := executor.Validate()
	assert.Equal(t, expectedErr, actualErr)
}

func TestExecutorRender(t *testing.T) {
	sampleCfgDoc := executorDoc(t, fmt.Sprintf(executorConfigTmpl, "init"))
	executor, err := executors.NewClusterctlExecutor(
		ifc.ExecutorConfig{
			ExecutorDocument: sampleCfgDoc,
			Helper:           makeDefaultHelper(t, "../../clusterctl/client/testdata", defaultMetadataPath),
		})
	require.NoError(t, err)
	actualOut := &bytes.Buffer{}
	actualErr := executor.Render(actualOut, ifc.RenderOptions{})
	assert.Equal(t, renderedDocs, actualOut.String())
	assert.NoError(t, actualErr)
}

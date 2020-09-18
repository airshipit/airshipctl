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

package client_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/runtime/schema"

	cctlclient "opendev.org/airship/airshipctl/pkg/clusterctl/client"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	airerrors "opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/testutil/fs"
)

var (
	executorConfigTmpl = `
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  name: clusterctl-v1
action: %s
init-options:
  core-provider: "cluster-api:v0.3.3"
  kubeConfigRef:
    apiVersion: airshipit.org/v1alpha1
    kind: KubeConfig
    name: sample-name
providers:
  - name: "cluster-api"
    type: "CoreProvider"
    versions:
      v0.3.3: manifests/function/capi/v0.3.3`
)

func TestRegisterExecutor(t *testing.T) {
	registry := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
	expectedGVK := schema.GroupVersionKind{
		Group:   "airshipit.org",
		Version: "v1alpha1",
		Kind:    "Clusterctl",
	}
	err := cctlclient.RegisterExecutor(registry)
	require.NoError(t, err)

	_, found := registry[expectedGVK]
	assert.True(t, found)
}

func TestNewExecutor(t *testing.T) {
	sampleCfgDoc := executorDoc(t, "init")
	bundle, err := document.NewBundleByPath("testdata/executor_init")
	require.NoError(t, err)

	testCases := []struct {
		name        string
		helper      ifc.Helper
		expectedErr error
	}{
		{
			name:   "New Clusterctl Executor",
			helper: makeDefaultHelper(t),
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			_, actualErr := cctlclient.NewExecutor(ifc.ExecutorConfig{
				ExecutorDocument: sampleCfgDoc,
				ExecutorBundle:   bundle,
				Helper:           tt.helper,
			})
			assert.Equal(t, tt.expectedErr, actualErr)
		})
	}
	require.NoError(t, err)
}

func TestExecutorRun(t *testing.T) {
	errTmpFile := errors.New("TmpFile error")

	testCases := []struct {
		name        string
		cfgDoc      document.Document
		fs          document.FileSystem
		bundlePath  string
		expectedEvt []events.Event
	}{
		{
			name:       "Error unknown action",
			cfgDoc:     executorDoc(t, "someAction"),
			bundlePath: "testdata/executor_init",
			expectedEvt: []events.Event{
				wrapError(cctlclient.ErrUnknownExecutorAction{Action: "someAction"}),
			},
		},
		{
			name:   "Error temporary file",
			cfgDoc: executorDoc(t, "init"),
			fs: fs.MockFileSystem{
				MockTempFile: func(string, string) (document.File, error) {
					return nil, errTmpFile
				},
			},
			bundlePath: "testdata/executor_init",
			expectedEvt: []events.Event{
				{
					Type: events.ClusterctlType,
					ClusterctlEvent: events.ClusterctlEvent{
						Operation: events.ClusterctlInitStart,
					},
				},
				wrapError(errTmpFile),
			},
		},
		{
			name:   "Regular Run init",
			cfgDoc: executorDoc(t, "init"),
			fs: fs.MockFileSystem{
				MockTempFile: func(string, string) (document.File, error) {
					return fs.TestFile{
						MockName:  func() string { return "filename" },
						MockWrite: func() (int, error) { return 0, nil },
						MockClose: func() error { return nil },
					}, nil
				},
				MockRemoveAll: func() error { return nil },
			},
			bundlePath: "testdata/executor_init",
			expectedEvt: []events.Event{
				{
					Type: events.ClusterctlType,
					ClusterctlEvent: events.ClusterctlEvent{
						Operation: events.ClusterctlInitStart,
					},
				},
				{
					Type: events.ClusterctlType,
					ClusterctlEvent: events.ClusterctlEvent{
						Operation: events.ClusterctlInitEnd,
					},
				},
			},
		},
		// TODO add move tests here
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := document.NewBundleByPath(tt.bundlePath)
			require.NoError(t, err)
			kubeCfg := kubeconfig.NewKubeConfig(
				kubeconfig.FromByte([]byte("someKubeConfig")),
				kubeconfig.InjectFileSystem(tt.fs),
			)
			executor, err := cctlclient.NewExecutor(
				ifc.ExecutorConfig{
					ExecutorDocument: tt.cfgDoc,
					ExecutorBundle:   bundle,
					Helper:           makeDefaultHelper(t),
					KubeConfig:       kubeCfg,
				})
			require.NoError(t, err)
			ch := make(chan events.Event)
			go executor.Run(ch, ifc.RunOptions{DryRun: true})
			var actualEvt []events.Event
			for evt := range ch {
				if evt.Type == events.ClusterctlType {
					// Set message to empty string, so it's not compared
					evt.ClusterctlEvent.Message = ""
				}
				actualEvt = append(actualEvt, evt)
			}
			assert.Equal(t, tt.expectedEvt, actualEvt)
		})
	}
}

func TestExecutorValidate(t *testing.T) {
	sampleCfgDoc := executorDoc(t, "init")
	bundle, err := document.NewBundleByPath("testdata/executor_init")
	require.NoError(t, err)
	executor, err := cctlclient.NewExecutor(
		ifc.ExecutorConfig{
			ExecutorDocument: sampleCfgDoc,
			ExecutorBundle:   bundle,
			Helper:           makeDefaultHelper(t),
		})
	require.NoError(t, err)
	expectedErr := airerrors.ErrNotImplemented{}
	actualErr := executor.Validate()
	assert.Equal(t, expectedErr, actualErr)
}
func TestExecutorRender(t *testing.T) {
	sampleCfgDoc := executorDoc(t, "init")
	bundle, err := document.NewBundleByPath("testdata/executor_init")
	require.NoError(t, err)
	expectedErr := airerrors.ErrNotImplemented{}

	executor, err := cctlclient.NewExecutor(
		ifc.ExecutorConfig{
			ExecutorDocument: sampleCfgDoc,
			ExecutorBundle:   bundle,
			Helper:           makeDefaultHelper(t),
		})
	require.NoError(t, err)
	actualOut := &bytes.Buffer{}
	actualErr := executor.Render(actualOut, ifc.RenderOptions{})
	assert.Equal(t, expectedErr, actualErr)
}

func makeDefaultHelper(t *testing.T) ifc.Helper {
	t.Helper()
	cfg := config.NewConfig()
	cfg.Manifests[config.AirshipDefaultManifest].TargetPath = "./testdata"
	cfg.Manifests[config.AirshipDefaultManifest].MetadataPath = "metadata.yaml"
	cfg.SetLoadedConfigPath(".")
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	return helper
}

func executorDoc(t *testing.T, action string) document.Document {
	cfg := []byte(fmt.Sprintf(executorConfigTmpl, action))
	cfgDoc, err := document.NewDocumentFromBytes(cfg)
	require.NoError(t, err)
	return cfgDoc
}

func wrapError(err error) events.Event {
	return events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}

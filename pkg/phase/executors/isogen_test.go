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
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/testutil"
	testcontainer "opendev.org/airship/airshipctl/testutil/container"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
)

var (
	isogenExecutorDoc = `
apiVersion: airshipit.org/v1alpha1
kind: IsoConfiguration
metadata:
  name: isogen
  labels:
    airshipit.org/deploy-k8s: "false"
builder:
  outputFileName: ephemeral.iso
container:
  containerRuntime: docker
  image: quay.io/airshipit/image-builder:latest-ubuntu_focal
  volume: /srv/images:/config`
	executorBundlePath = "../../bootstrap/isogen/testdata/primary/site/test-site/ephemeral/bootstrap"
)

func TestNewIsogenExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(isogenExecutorDoc))
	require.NoError(t, err)
	_, err = executors.NewIsogenExecutor(ifc.ExecutorConfig{
		ExecutorDocument: execDoc,
		BundleFactory:    testBundleFactory(executorBundlePath)})
	require.NoError(t, err)
}

func TestIsogenExecutorRun(t *testing.T) {
	bundle, err := document.NewBundleByPath(executorBundlePath)
	require.NoError(t, err)
	require.NotNil(t, bundle)

	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	emptyFileName := filepath.Join(tempVol, "ephemeral.iso")
	emptyFile, createErr := os.Create(emptyFileName)
	require.NoError(t, createErr)
	log.Println(emptyFile)
	emptyFile.Close()

	volBind := tempVol + ":/dst"
	testCfg := &v1alpha1.IsoConfiguration{
		IsoContainer: &v1alpha1.IsoContainer{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Isogen: &v1alpha1.Isogen{
			OutputFileName: "ephemeral.iso",
		},
	}
	testDoc := &testdoc.MockDocument{
		MockAsYAML: func() ([]byte, error) { return []byte("TESTDOC"), nil },
	}

	testCases := []struct {
		name        string
		builder     *testcontainer.MockContainer
		expectedEvt []events.Event
	}{
		{
			name: "Run isogen successfully",
			builder: &testcontainer.MockContainer{
				MockRunCommand:        func() error { return nil },
				MockGetID:             func() string { return "TESTID" },
				MockRmContainer:       func() error { return nil },
				MockWaitUntilFinished: func() error { return nil },
			},
			expectedEvt: []events.Event{
				events.NewEvent().WithIsogenEvent(events.IsogenEvent{
					Operation: events.IsogenStart,
				}),
				events.NewEvent().WithIsogenEvent(events.IsogenEvent{
					Operation: events.IsogenValidation,
				}),
				events.NewEvent().WithIsogenEvent(events.IsogenEvent{
					Operation: events.IsogenEnd,
				}),
			},
		},
		{
			name: "Fail on container command",
			builder: &testcontainer.MockContainer{
				MockRunCommand: func() error {
					return container.ErrRunContainerCommand{Cmd: "super fail"}
				},
				MockGetID:       func() string { return "TESTID" },
				MockRmContainer: func() error { return nil },
			},

			expectedEvt: []events.Event{
				events.NewEvent().WithIsogenEvent(events.IsogenEvent{
					Operation: events.IsogenStart,
				}),
				wrapError(container.ErrRunContainerCommand{Cmd: "super fail"}),
			},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			executor := &executors.IsogenExecutor{
				ExecutorDocument: testDoc,
				ExecutorBundle:   bundle,
				ImgConf:          testCfg,
				Builder:          tt.builder,
			}
			ch := make(chan events.Event)
			go executor.Run(ch, ifc.RunOptions{})
			var actualEvt []events.Event
			for evt := range ch {
				// Skip timestamp for comparison
				evt.Timestamp = time.Time{}
				if evt.Type == events.IsogenType {
					// Set message to empty string, so it's not compared
					evt.IsogenEvent.Message = ""
				}
				actualEvt = append(actualEvt, evt)
			}
			for i := range tt.expectedEvt {
				// Skip timestamp for comparison
				tt.expectedEvt[i].Timestamp = time.Time{}
			}
			assert.Equal(t, tt.expectedEvt, actualEvt)
		})
	}
}

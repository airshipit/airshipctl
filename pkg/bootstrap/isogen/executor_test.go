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

package isogen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/testutil"
)

var (
	executorDoc = `
apiVersion: airshipit.org/v1alpha1
kind: ImageConfiguration
metadata:
  name: isogen
  labels:
    airshipit.org/deploy-k8s: "false"
builder:
  networkConfigFileName: network-config
  outputMetadataFileName: output-metadata.yaml
  userDataFileName: user-data
container:
  containerRuntime: docker
  image: quay.io/airshipit/isogen:latest-ubuntu_focal
  volume: /srv/iso:/config`
	executorBundlePath = "testdata/primary/site/test-site/ephemeral/bootstrap"
)

func TestRegisterExecutor(t *testing.T) {
	registry := make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
	expectedGVK := schema.GroupVersionKind{
		Group:   "airshipit.org",
		Version: "v1alpha1",
		Kind:    "ImageConfiguration",
	}
	err := RegisterExecutor(registry)
	require.NoError(t, err)

	_, found := registry[expectedGVK]
	assert.True(t, found)
}

func TestNewExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(executorDoc))
	require.NoError(t, err)
	_, err = NewExecutor(ifc.ExecutorConfig{
		ExecutorDocument: execDoc,
		BundleFactory:    testBundleFactory(executorBundlePath)})
	require.NoError(t, err)
}

func TestExecutorRun(t *testing.T) {
	bundle, err := document.NewBundleByPath(executorBundlePath)
	require.NoError(t, err)
	require.NotNil(t, bundle)

	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	volBind := tempVol + ":/dst"
	testCfg := &v1alpha1.ImageConfiguration{
		Container: &v1alpha1.Container{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Builder: &v1alpha1.Builder{
			UserDataFileName:      "user-data",
			NetworkConfigFileName: "net-conf",
		},
	}
	testDoc := &MockDocument{
		MockAsYAML: func() ([]byte, error) { return []byte("TESTDOC"), nil },
	}

	testCases := []struct {
		name        string
		builder     *mockContainer
		expectedEvt []events.Event
	}{
		{
			name: "Run isogen successfully",
			builder: &mockContainer{
				runCommand:        func() error { return nil },
				getID:             func() string { return "TESTID" },
				rmContainer:       func() error { return nil },
				waitUntilFinished: func() error { return nil },
			},
			expectedEvt: []events.Event{
				{
					Type: events.IsogenType,
					IsogenEvent: events.IsogenEvent{
						Operation: events.IsogenStart,
					},
				},
				{
					Type: events.IsogenType,
					IsogenEvent: events.IsogenEvent{
						Operation: events.IsogenValidation,
					},
				},
				{
					Type: events.IsogenType,
					IsogenEvent: events.IsogenEvent{
						Operation: events.IsogenEnd,
					},
				},
			},
		},
		{
			name: "Fail on container command",
			builder: &mockContainer{
				runCommand: func() error {
					return container.ErrRunContainerCommand{Cmd: "super fail"}
				},
				getID:       func() string { return "TESTID" },
				rmContainer: func() error { return nil },
			},

			expectedEvt: []events.Event{
				{
					Type: events.IsogenType,
					IsogenEvent: events.IsogenEvent{
						Operation: events.IsogenStart,
					},
				},
				wrapError(container.ErrRunContainerCommand{Cmd: "super fail"}),
			},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{
				ExecutorDocument: testDoc,
				ExecutorBundle:   bundle,
				imgConf:          testCfg,
				builder:          tt.builder,
			}
			ch := make(chan events.Event)
			go executor.Run(ch, ifc.RunOptions{})
			var actualEvt []events.Event
			for evt := range ch {
				if evt.Type == events.IsogenType {
					// Set message to empty string, so it's not compared
					evt.IsogenEvent.Message = ""
				}
				actualEvt = append(actualEvt, evt)
			}
			assert.Equal(t, tt.expectedEvt, actualEvt)
		})
	}
}

func wrapError(err error) events.Event {
	return events.Event{
		Type: events.ErrorType,
		ErrorEvent: events.ErrorEvent{
			Error: err,
		},
	}
}

func testBundleFactory(path string) document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		return document.NewBundleByPath(path)
	}
}

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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/ephemeral"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/testutil"
	testcontainer "opendev.org/airship/airshipctl/testutil/container"
)

var (
	executorEphemeralDoc = `
apiVersion: airshipit.org/v1alpha1
kind: BootConfiguration
metadata:
  name: ephemeral-az-genesis
  labels:
    airshipit.org/deploy-k8s: "false"
ephemeralCluster:
  bootstrapCommand: create
  configFilename: azure-config.yaml
bootstrapContainer:
  containerRuntime: docker
  image: quay.io/sshiba/capz-bootstrap:latest
  volume: /home/esidshi/.airship:/kube
  saveKubeconfigFileName: capz.kubeconfig
`
)

// TestNewEphemeralExecutor - Unit testing function NewExecutor()
func TestNewEphemeralExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(executorEphemeralDoc))
	require.NoError(t, err)
	_, err = executors.NewEphemeralExecutor(ifc.ExecutorConfig{ExecutorDocument: execDoc})
	require.NoError(t, err)
}

// TestExecutorEphemeralRun - Unit testing method (c *Executor) Run()
func TestExecutorEphemeralRun(t *testing.T) {
	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	volBind := tempVol + ":/dst"
	configFilename := "dummy-config.yaml"
	configPath := filepath.Join(tempVol, configFilename)

	assert.NoError(t, testConfigFile(configPath),
		ephemeral.ErrBootstrapContainerRun{},
		"Failed to create dummy config file")

	testCfg := &v1alpha1.BootConfiguration{
		BootstrapContainer: v1alpha1.BootstrapContainer{
			Volume:           volBind,
			ContainerRuntime: "docker",
			Image:            "quay.io/sshiba/capz-bootstrap:latest",
			Kubeconfig:       "dummy.kubeconfig",
		},
		EphemeralCluster: v1alpha1.EphemeralCluster{
			BootstrapCommand: "create",
			ConfigFilename:   configFilename,
		},
	}

	testCases := []struct {
		name        string
		container   *testcontainer.MockContainer
		expectedErr error
	}{
		{
			name: "Run bootstrap container successfully",
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return nil },
				MockRmContainer: func() error { return nil },
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = "exited"
					state.ExitCode = 0
					return state, nil
				},
			},
		},
		{
			name: "Run bootstrap container with Unknown Error",
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return ephemeral.ErrBootstrapContainerRun{} },
				MockRmContainer: func() error { return nil },
			},
			expectedErr: ephemeral.ErrBootstrapContainerRun{},
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			executor := &executors.EphemeralExecutor{
				BootConf:  testCfg,
				Container: tt.container,
			}
			err := executor.Run(ifc.RunOptions{})
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func testConfigFile(path string) error {
	_, err := os.Create(path)
	return err
}

// TestValidate - Unit testing function Validate()
func TestEphemeralValidate(t *testing.T) {
	executor := &executors.EphemeralExecutor{}

	actualErr := executor.Validate()
	assert.Equal(t, errors.ErrNotImplemented{}, actualErr)
}

// TestEphemeralRender - Unit testing function Render()
func TestEphemeralRender(t *testing.T) {
	executor := &executors.EphemeralExecutor{}

	writerReader := bytes.NewBuffer([]byte{})
	actualErr := executor.Render(writerReader, ifc.RenderOptions{})
	assert.Equal(t, nil, actualErr)
}

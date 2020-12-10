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

package ephemeral_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	api "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/ephemeral"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/log"
	testcontainer "opendev.org/airship/airshipctl/testutil/container"
)

// Unit test for the method getContainerStatus()
func TestGetContainerStatus(t *testing.T) {
	testCfg := &api.BootConfiguration{
		BootstrapContainer: api.BootstrapContainer{
			ContainerRuntime: "docker",
			Image:            "quay.io/dummy/dummy:latest",
			Volume:           "/dummy:/dummy",
			Kubeconfig:       "dummy.kubeconfig",
		},
		EphemeralCluster: api.EphemeralCluster{
			BootstrapCommand: "dummy",
			ConfigFilename:   "dummy.yaml",
		},
	}

	tests := []struct {
		container      *testcontainer.MockContainer
		cfg            *api.BootConfiguration
		debug          bool
		expectedStatus container.Status
		expectedErr    error
	}{
		{
			// options.Container.InspectContainer() returning error
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = ephemeral.BootNullString
					state.ExitCode = 0
					return state, ephemeral.ErrInvalidBootstrapCommand{}
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: ephemeral.BootNullString,
			expectedErr:    ephemeral.ErrInvalidBootstrapCommand{},
		},
		{
			// options.Container.GetContainerLogs() returns error
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.RunningContainerStatus
					state.ExitCode = 1
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, ephemeral.ErrInvalidBootstrapCommand{}
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: ephemeral.BootNullString,
			expectedErr:    ephemeral.ErrInvalidBootstrapCommand{},
		},
		{
			// Container running and no errors
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.RunningContainerStatus
					state.ExitCode = 0
					return state, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.RunningContainerStatus,
			expectedErr:    nil,
		},
		{
			// Container running and with ContainerLoadEphemeralConfigError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 1
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 1,
				ErrMsg:   ephemeral.ContainerLoadEphemeralConfigError,
			},
		},
		{
			// Container running and with ContainerValidationEphemeralConfigError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 2
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 2,
				ErrMsg:   ephemeral.ContainerValidationEphemeralConfigError,
			},
		},
		{
			// Container running and with ContainerSetEnvVarsError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 3
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 3,
				ErrMsg:   ephemeral.ContainerSetEnvVarsError,
			},
		},
		{
			// Container running and with ContainerUnknownCommandError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 4
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 4,
				ErrMsg:   ephemeral.ContainerUnknownCommandError,
			},
		},
		{
			// Container running and with ContainerCreationEphemeralFailedError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 5
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 5,
				ErrMsg:   ephemeral.ContainerCreationEphemeralFailedError,
			},
		},
		{
			// Container running and with ContainerDeletionEphemeralFailedError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 6
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 6,
				ErrMsg:   ephemeral.ContainerDeletionEphemeralFailedError,
			},
		},
		{
			// Container running and with ContainerHelpCommandFailedError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 7
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 7,
				ErrMsg:   ephemeral.ContainerHelpCommandFailedError,
			},
		},
		{
			// Container running and with ContainerUnknownError
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 8
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:            testCfg,
			debug:          false,
			expectedStatus: container.ExitedContainerStatus,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 8,
				ErrMsg:   ephemeral.ContainerUnknownError,
			},
		},
	}

	for _, tt := range tests {
		outBuf := &bytes.Buffer{}
		log.Init(tt.debug, outBuf)
		bootstrapOpts := ephemeral.BootstrapContainerOptions{
			Container: tt.container,
			Cfg:       tt.cfg,
			Sleep:     func(_ time.Duration) {},
			Debug:     tt.debug,
		}
		actualStatus, actualErr := bootstrapOpts.GetContainerStatus()
		assert.Equal(t, tt.expectedStatus, actualStatus)
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

// Unit test for the method waitUntilContainerExitsOrTimesout()
func TestWaitUntilContainerExitsOrTimesout(t *testing.T) {
	testCfg := &api.BootConfiguration{
		BootstrapContainer: api.BootstrapContainer{
			ContainerRuntime: "docker",
			Image:            "quay.io/dummy/dummy:latest",
			Volume:           "/dummy:/dummy",
			Kubeconfig:       "dummy.kubeconfig",
		},
		EphemeralCluster: api.EphemeralCluster{
			BootstrapCommand: "dummy",
			ConfigFilename:   "dummy.yaml",
		},
	}

	tests := []struct {
		container   *testcontainer.MockContainer
		cfg         *api.BootConfiguration
		debug       bool
		maxRetries  int
		expectedErr error
	}{
		{
			// Test: container exits normally
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 0
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:         testCfg,
			debug:       false,
			maxRetries:  10,
			expectedErr: nil,
		},
		{
			// Test: container times out
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.RunningContainerStatus
					state.ExitCode = 0
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:         testCfg,
			debug:       false,
			maxRetries:  1,
			expectedErr: ephemeral.ErrNumberOfRetriesExceeded{},
		},
		{
			// Test: options.GetContainerStatus() returns error
			container: &testcontainer.MockContainer{
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = ephemeral.BootNullString
					state.ExitCode = 0
					return state, ephemeral.ErrInvalidBootstrapCommand{}
				},
			},
			cfg:         testCfg,
			debug:       false,
			maxRetries:  10,
			expectedErr: ephemeral.ErrInvalidBootstrapCommand{},
		},
	}
	for _, tt := range tests {
		outBuf := &bytes.Buffer{}
		log.Init(tt.debug, outBuf)
		bootstrapOpts := ephemeral.BootstrapContainerOptions{
			Container: tt.container,
			Cfg:       tt.cfg,
			Sleep:     func(_ time.Duration) {},
			Debug:     tt.debug,
		}
		actualErr := bootstrapOpts.WaitUntilContainerExitsOrTimesout(tt.maxRetries, "dummy-config.yaml", "dummy")
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

// Unit test for the method createBootstrapContainer()
func TestCreateBootstrapContainer(t *testing.T) {
	testCfg := &api.BootConfiguration{
		BootstrapContainer: api.BootstrapContainer{
			ContainerRuntime: "docker",
			Image:            "quay.io/dummy/dummy:latest",
			Volume:           "/dummy:/dummy",
			Kubeconfig:       "dummy.kubeconfig",
		},
		EphemeralCluster: api.EphemeralCluster{
			BootstrapCommand: "dummy",
			ConfigFilename:   "dummy.yaml",
		},
	}

	tests := []struct {
		container        *testcontainer.MockContainer
		cfg              *api.BootConfiguration
		debug            bool
		bootstrapCommand string
		expectedErr      error
	}{
		{
			// Test: Create Ephemeral Cluster successful
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return nil },
				MockRmContainer: func() error { return nil },
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 0
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:              testCfg,
			bootstrapCommand: ephemeral.BootCmdCreate,
			debug:            false,
			expectedErr:      nil,
		},
		{
			// Test: Delete Ephemeral Cluster successful
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return nil },
				MockRmContainer: func() error { return nil },
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 0
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:              testCfg,
			bootstrapCommand: ephemeral.BootCmdDelete,
			debug:            false,
			expectedErr:      nil,
		},
		{
			// Test: Create Ephemeral Cluster exit with error
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return nil },
				MockRmContainer: func() error { return nil },
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 1
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:              testCfg,
			bootstrapCommand: ephemeral.BootCmdCreate,
			debug:            false,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 1,
				ErrMsg:   ephemeral.ContainerLoadEphemeralConfigError,
			},
		},
		{
			// Test: Delete Ephemeral Cluster exit with error
			container: &testcontainer.MockContainer{
				MockRunCommand:  func() error { return nil },
				MockRmContainer: func() error { return nil },
				MockInspectContainer: func() (container.State, error) {
					state := container.State{}
					state.Status = container.ExitedContainerStatus
					state.ExitCode = 1
					return state, nil
				},
				MockGetContainerLogs: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			cfg:              testCfg,
			bootstrapCommand: ephemeral.BootCmdDelete,
			debug:            false,
			expectedErr: ephemeral.ErrBootstrapContainerRun{
				ExitCode: 1,
				ErrMsg:   ephemeral.ContainerLoadEphemeralConfigError,
			},
		},
		{
			// Test: options.Container.RunCommand() returns error
			container: &testcontainer.MockContainer{
				MockRunCommand: func() error { return ephemeral.ErrBootstrapContainerRun{} },
			},
			cfg:              testCfg,
			bootstrapCommand: ephemeral.BootCmdCreate,
			debug:            false,
			expectedErr:      ephemeral.ErrBootstrapContainerRun{},
		},
	}

	for _, tt := range tests {
		tt.cfg.EphemeralCluster.BootstrapCommand = tt.bootstrapCommand
		bootstrapOpts := ephemeral.BootstrapContainerOptions{
			Container: tt.container,
			Cfg:       tt.cfg,
			Sleep:     func(_ time.Duration) {},
			Debug:     tt.debug,
		}
		actualErr := bootstrapOpts.CreateBootstrapContainer()
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

func TestVerifyInputs(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *api.BootConfiguration
		expectedErr error
	}{
		{
			name: "Verify successful",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "/dummy:/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Missing Ephemeral cluster config file",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "/dummy:/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "",
				},
			},
			expectedErr: ephemeral.ErrInvalidInput{
				What: ephemeral.MissingConfigError,
			},
		},
		{
			name: "Missing Volume mount for the Bootstrap container",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: ephemeral.ErrInvalidInput{
				What: ephemeral.MissingVolumeError,
			},
		},
		{
			name: "Bootstrap container Volume mount mal formed 1",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: nil,
		},
		{
			name: "Bootstrap container Volume mount mal formed 2",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "/dummy:/dummy:/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: ephemeral.ErrVolumeMalFormed{},
		},
		{
			name: "Bootstrap container image missing",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "docker",
					Image:            "",
					Volume:           "/dummy:/dummy:/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: ephemeral.ErrInvalidInput{
				What: ephemeral.MissingContainerImageError,
			},
		},
		{
			name: "Bootstrap container runtime missing",
			cfg: &api.BootConfiguration{
				BootstrapContainer: api.BootstrapContainer{
					ContainerRuntime: "",
					Image:            "quay.io/dummy/dummy:latest",
					Volume:           "/dummy:/dummy:/dummy",
					Kubeconfig:       "dummy.kubeconfig",
				},
				EphemeralCluster: api.EphemeralCluster{
					BootstrapCommand: "dummy",
					ConfigFilename:   "dummy.yaml",
				},
			},
			expectedErr: ephemeral.ErrInvalidInput{
				What: ephemeral.MissingContainerRuntimeError,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		bootstrapOpts := ephemeral.BootstrapContainerOptions{
			Cfg: tt.cfg,
		}
		t.Run(tt.name, func(subTest *testing.T) {
			actualErr := bootstrapOpts.VerifyInputs()
			assert.Equal(subTest, tt.expectedErr, actualErr)
		})
	}
}

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

package container

import (
	"context"
	"io"
)

const (
	// DriverDocker indicates that docker driver should be used in container constructor
	DriverDocker = "docker"
)

// Status type provides container status
type Status string

// State provides information about the Container
type State struct {
	// ExitCode: returns the container's exit code. Zero means exited normally, otherwise errored
	ExitCode int
	// Status: String representation of the container state.
	Status Status
}

// Container interface abstraction for container.
// Particular implementation depends on container runtime environment (CRE). Interface
// defines methods that must be implemented for CRE (e.g. docker, containerd or CRI-O)
type Container interface {
	ImagePull() error
	RunCommand(RunCommandOptions) error
	GetContainerLogs(GetLogOptions) (io.ReadCloser, error)
	InspectContainer() (State, error)
	WaitUntilFinished() error
	RmContainer() error
	GetID() string
}

// RunCommandOptions options for RunCommand
type RunCommandOptions struct {
	Privileged  bool
	HostNetwork bool

	Cmd     []string
	EnvVars []string
	Binds   []string

	Mounts []Mount
	Input  io.Reader
}

// Mount describes mount settings
type Mount struct {
	ReadOnly bool
	Type     string
	Dst      string
	Src      string
}

// GetLogOptions options for getting logs
// If both Stderr and Stdout are specified the logs will contain both stderr and stdout
type GetLogOptions struct {
	Stderr bool
	Stdout bool
	Follow bool
}

// NewContainer returns instance of Container interface implemented by particular driver
// Returned instance type (i.e. implementation) depends on driver specified via function
// arguments (e.g. "docker").
// Supported drivers:
//   * docker
func NewContainer(ctx context.Context, driver string, url string) (Container, error) {
	switch driver {
	case "":
		return nil, ErrNoContainerDriver{}
	case DriverDocker:
		cli, err := NewDockerClient(ctx)
		if err != nil {
			return nil, err
		}
		return NewDockerContainer(ctx, url, cli)
	default:
		return nil, ErrContainerDrvNotSupported{Driver: driver}
	}
}

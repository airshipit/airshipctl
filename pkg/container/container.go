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

// Container interface abstraction for container.
// Particular implementation depends on container runtime environment (CRE). Interface
// defines methods that must be implemented for CRE (e.g. docker, containerd or CRI-O)
type Container interface {
	ImagePull() error
	RunCommand([]string, io.Reader, []string, []string) error
	GetContainerLogs() (io.ReadCloser, error)
	RmContainer() error
	GetID() string
	WaitUntilFinished() error
}

// NewContainer returns instance of Container interface implemented by particular driver
// Returned instance type (i.e. implementation) depends on driver specified via function
// arguments (e.g. "docker").
// Supported drivers:
//   * docker
func NewContainer(ctx *context.Context, driver string, url string) (Container, error) {
	switch driver {
	case "":
		return nil, ErrNoContainerDriver{}
	case "docker":
		cli, err := NewDockerClient(ctx)
		if err != nil {
			return nil, err
		}
		return NewDockerContainer(ctx, url, cli)
	default:
		return nil, ErrContainerDrvNotSupported{Driver: driver}
	}
}

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
	RunCommand([]string, io.Reader, []string, []string, bool) error
	RunCommandOutput([]string, io.Reader, []string, []string) (io.ReadCloser, error)
	RmContainer() error
	GetId() string
}

// NewContainer returns instance of Container interface implemented by particular driver
// Returned instance type (i.e. implementation) depends on driver specified via function
// arguments (e.g. "docker").
// Supported drivers:
//   * docker
func NewContainer(ctx *context.Context, driver string, url string) (Container, error) {
	switch driver {
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

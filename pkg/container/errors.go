package container

import (
	"fmt"
)

// ErrEmptyImageList returned if no image defined in filter found
type ErrEmptyImageList struct {
}

func (e ErrEmptyImageList) Error() string {
	return "Image List Error. No images filetered"
}

// ErrRunContainerCommand returned if container command
// exited with non-zero code
type ErrRunContainerCommand struct {
	Cmd string
}

func (e ErrRunContainerCommand) Error() string {
	return fmt.Sprintf("Command error. run '%s' for details", e.Cmd)
}

// ErrContainerDrvNotSupported returned if desired CRI is not supported
type ErrContainerDrvNotSupported struct {
	Driver string
}

func (e ErrContainerDrvNotSupported) Error() string {
	return fmt.Sprintf("Driver %s is not supported", e.Driver)
}

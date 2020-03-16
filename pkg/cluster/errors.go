package cluster

import "fmt"

// ErrInvalidStatusCheck denotes that something went wrong while handling a
// status-check annotation.
type ErrInvalidStatusCheck struct {
	What string
}

func (err ErrInvalidStatusCheck) Error() string {
	return fmt.Sprintf("invalid status-check: %s", err.What)
}

// ErrResourceNotFound is used when a resource is requrested from a StatusMap,
// but that resource can't be found
type ErrResourceNotFound struct {
	Resource string
}

func (err ErrResourceNotFound) Error() string {
	return fmt.Sprintf("could not find a status for resource %q", err.Resource)
}

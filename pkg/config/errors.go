package config

import (
	"fmt"
)

// ErrBootstrapInfoNotFound returned if bootstrap
// information is not found for cluster
type ErrBootstrapInfoNotFound struct {
	Name string
}

func (e ErrBootstrapInfoNotFound) Error() string {
	return fmt.Sprintf("Bootstrap info %s not found", e.Name)
}

package config

import (
	"fmt"
	"strings"
)

// Repo errors

// ErrMutuallyExclusiveAuthSSHPass is returned when ssh-pass type
// is selected and http-pass, ssh-key or key-pass options are defined
type ErrIncompatibleAuthOptions struct {
	ForbiddenOptions []string
	AuthType         string
}

func NewErrIncompetibleAuthOptions(fo []string, ao string) error {
	return ErrIncompatibleAuthOptions{
		ForbiddenOptions: fo,
		AuthType:         ao,
	}
}

func (e ErrIncompatibleAuthOptions) Error() string {
	return fmt.Sprintf("Can not use %s options, with auth type %s", e.ForbiddenOptions, e.AuthType)
}

// ErrAuthTypeNotSupported is returned with wrong AuthType is provided
type ErrAuthTypeNotSupported struct {
}

func (e ErrAuthTypeNotSupported) Error() string {
	return "Invalid auth, allowed types: " + strings.Join(AllowedAuthTypes, ",")
}

// ErrRepoSpecRequiresURL is returned when repository URL is not specified
type ErrRepoSpecRequiresURL struct {
}

func (e ErrRepoSpecRequiresURL) Error() string {
	return "Repostory spec requires url"
}

// ErrMutuallyExclusiveCheckout is returned if
// mutually exclusive options are given as checkout options
type ErrMutuallyExclusiveCheckout struct {
}

func (e ErrMutuallyExclusiveCheckout) Error() string {
	return "Chekout mutually execlusive, use either: commit-hash, branch or tag"
}

// ErrBootstrapInfoNotFound returned if bootstrap
// information is not found for cluster
type ErrBootstrapInfoNotFound struct {
	Name string
}

func (e ErrBootstrapInfoNotFound) Error() string {
	return fmt.Sprintf("Bootstrap info %q not found", e.Name)
}

// ErrWrongConfig returned in case of incorrect configuration
type ErrWrongConfig struct {
}

func (e ErrWrongConfig) Error() string {
	return "Wrong configuration"
}

// ErrMissingConfig returned in case of missing configuration
type ErrMissingConfig struct {
	What string
}

func (e ErrMissingConfig) Error() string {
	return "Missing configuration: " + e.What
}

// ErrConfigFailed returned in case of failure during configuration
type ErrConfigFailed struct {
}

func (e ErrConfigFailed) Error() string {
	return "Configuration failed to complete."
}

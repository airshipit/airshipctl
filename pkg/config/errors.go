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

package config

import (
	"fmt"
	"strings"
)

// ErrIncompatibleAuthOptions is returned when incompatible
// AuthTypes are provided
type ErrIncompatibleAuthOptions struct {
	ForbiddenOptions []string
	AuthType         string
}

// NewErrIncompetibleAuthOptions returns Error of type
// ErrIncompatibleAuthOptions
func NewErrIncompetibleAuthOptions(fo []string, ao string) error {
	return ErrIncompatibleAuthOptions{
		ForbiddenOptions: fo,
		AuthType:         ao,
	}
}

// Error of type ErrIncompatibleAuthOptions is returned when
// ssh-pass type is selected and http-pass, ssh-key or key-pass
// options are defined
func (e ErrIncompatibleAuthOptions) Error() string {
	return fmt.Sprintf("Can not use %s options, with auth type %s", e.ForbiddenOptions, e.AuthType)
}

// ErrAuthTypeNotSupported is returned when wrong AuthType is provided
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

// ErrInvalidConfig returned in case of incorrect configuration
type ErrInvalidConfig struct {
	What string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("Invalid configuration: %s", e.What)
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

// ErrMissingCurrentContext returned in case --current used without setting current-context
type ErrMissingCurrentContext struct {
}

func (e ErrMissingCurrentContext) Error() string {
	return "Current context must be set before using --current flag"
}

// ErrMissingManagementConfiguration means the management configuration was not defined for the active cluster.
type ErrMissingManagementConfiguration struct {
	cluster *Cluster
}

func (e ErrMissingManagementConfiguration) Error() string {
	return fmt.Sprintf("Management configuration %s for cluster %s undefined.", e.cluster.ManagementConfiguration,
		e.cluster.NameInKubeconf)
}

// ErrMissingPrimaryRepo returned when Primary Repository is not set in context manifest
type ErrMissingPrimaryRepo struct {
}

func (e ErrMissingPrimaryRepo) Error() string {
	return "Current context manifest must have primary repository set"
}

// ErrConflictingAuthOptions returned in case both token and username/password is set at same time
type ErrConflictingAuthOptions struct {
}

func (e ErrConflictingAuthOptions) Error() string {
	return "you cannot specify token and username/password at the same time"
}

// ErrConflictingClusterOptions returned when both certificate-authority and
// insecure-skip-tls-verify is set at same time
type ErrConflictingClusterOptions struct {
}

func (e ErrConflictingClusterOptions) Error() string {
	return "you cannot specify a certificate-authority and insecure-skip-tls-verify mode at the same time"
}

// ErrEmptyClusterName returned when empty cluster name is set
type ErrEmptyClusterName struct {
}

func (e ErrEmptyClusterName) Error() string {
	return "you must specify a non-empty cluster name"
}

// ErrConflictingContextOptions returned when both context and --current is set at same time
type ErrConflictingContextOptions struct {
}

func (e ErrConflictingContextOptions) Error() string {
	return "you cannot specify context and --current Flag at the same time"
}

// ErrEmptyContextName returned when empty context name is set
type ErrEmptyContextName struct {
}

func (e ErrEmptyContextName) Error() string {
	return "you must specify a non-empty context name"
}

// ErrDecodingCredentials returned when the given string cannot be decoded
type ErrDecodingCredentials struct {
	Given string
}

func (e ErrDecodingCredentials) Error() string {
	return fmt.Sprintf("Error decoding credentials. String '%s' cannot not be decoded", e.Given)
}

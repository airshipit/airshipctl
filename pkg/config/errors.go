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

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	redfishdell "opendev.org/airship/airshipctl/pkg/remote/redfish/vendors/dell"
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
	return fmt.Sprintf("Cannot use %s options with an auth type %s", e.ForbiddenOptions, e.AuthType)
}

// ErrAuthTypeNotSupported is returned when wrong AuthType is provided
type ErrAuthTypeNotSupported struct {
}

func (e ErrAuthTypeNotSupported) Error() string {
	return "Invalid Auth type. Allowed types: " + strings.Join(AllowedAuthTypes, ",")
}

// ErrRepoSpecRequiresURL is returned when repository URL is not specified
type ErrRepoSpecRequiresURL struct {
}

func (e ErrRepoSpecRequiresURL) Error() string {
	return "Repository spec requires url."
}

// ErrMutuallyExclusiveCheckout is returned if
// mutually exclusive options are given as checkout options
type ErrMutuallyExclusiveCheckout struct {
}

func (e ErrMutuallyExclusiveCheckout) Error() string {
	return "Checkout mutually exclusive, use either: commit-hash, branch or tag."
}

// ErrRepositoryNotFound is returned if repository is empty
// when using in set-manifest
type ErrRepositoryNotFound struct {
	Name string
}

func (e ErrRepositoryNotFound) Error() string {
	return fmt.Sprintf("Repository %q not found.", e.Name)
}

// ErrMissingRepositoryName is returned if repository name is empty
// when using in set-manifest
type ErrMissingRepositoryName struct {
}

func (e ErrMissingRepositoryName) Error() string {
	return "Missing repository name."
}

// ErrMissingRepoURL is returned if repository is empty
// when using --primary in set-manifest
type ErrMissingRepoURL struct {
}

func (e ErrMissingRepoURL) Error() string {
	return "A valid URL should be specified."
}

// ErrMissingRepoCheckoutOptions is returned if repository checkout
// options is empty in set-manifest
type ErrMissingRepoCheckoutOptions struct {
}

func (e ErrMissingRepoCheckoutOptions) Error() string {
	return "Missing repository checkout options."
}

// ErrBootstrapInfoNotFound returned if bootstrap
// information is not found for cluster
type ErrBootstrapInfoNotFound struct {
	Name string
}

func (e ErrBootstrapInfoNotFound) Error() string {
	return fmt.Sprintf("Bootstrap info %q not found.", e.Name)
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

// ErrManagementConfigurationNotFound describes a situation in which a user has attempted to reference a management
// configuration that cannot be referenced.
type ErrManagementConfigurationNotFound struct {
	Name string
}

func (e ErrManagementConfigurationNotFound) Error() string {
	return fmt.Sprintf("Unknown management configuration '%s'.", e.Name)
}

// ErrMissingCurrentContext returned in case --current used without setting current-context
type ErrMissingCurrentContext struct {
}

func (e ErrMissingCurrentContext) Error() string {
	return "Current context must be set before using --current flag."
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
	return "Current context manifest must have a primary repository set."
}

// ErrMissingPhaseDocument returned when appropriate Phase document was not found in the filesystem
type ErrMissingPhaseDocument struct {
	PhaseName string
}

func (e ErrMissingPhaseDocument) Error() string {
	return fmt.Sprintf("Phase document '%s' was not found. "+
		"You can initialize it using 'airshipctl document init %s' command.", e.PhaseName, e.PhaseName)
}

// ErrConflictingAuthOptions returned in case both token and username/password is set at same time
type ErrConflictingAuthOptions struct {
}

func (e ErrConflictingAuthOptions) Error() string {
	return "Specifying token and username/password is not allowed at the same time."
}

// ErrConflictingClusterOptions returned when both certificate-authority and
// insecure-skip-tls-verify is set at same time
type ErrConflictingClusterOptions struct {
}

func (e ErrConflictingClusterOptions) Error() string {
	return "Specifying certificate-authority and insecure-skip-tls-verify mode is not allowed at the same time."
}

// ErrEmptyClusterName returned when empty cluster name is set
type ErrEmptyClusterName struct {
}

func (e ErrEmptyClusterName) Error() string {
	return "Cluster name must not be empty."
}

// ErrConflictingContextOptions returned when both context and --current is set at same time
type ErrConflictingContextOptions struct {
}

func (e ErrConflictingContextOptions) Error() string {
	return "Specifying context and --current Flag is not allowed at the same time."
}

// ErrEmptyContextName returned when empty context name is set
type ErrEmptyContextName struct {
}

func (e ErrEmptyContextName) Error() string {
	return "Context name must not be empty."
}

// ErrDecodingCredentials returned when the given string cannot be decoded
type ErrDecodingCredentials struct {
	Given string
}

func (e ErrDecodingCredentials) Error() string {
	return fmt.Sprintf("Error decoding credentials. String '%s' cannot not be decoded", e.Given)
}

// ErrUnknownManagementType describes a situation in which an unknown management type is listed in the airshipctl
// config.
type ErrUnknownManagementType struct {
	Type string
}

func (e ErrUnknownManagementType) Error() string {
	return fmt.Sprintf("Unknown management type '%s'. Known types include '%s' and '%s'.", e.Type,
		redfish.ClientType, redfishdell.ClientType)
}

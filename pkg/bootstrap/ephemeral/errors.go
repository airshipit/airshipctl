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

package ephemeral

import "fmt"

const (
	// ContainerLoadEphemeralConfigError ...
	ContainerLoadEphemeralConfigError = "Container error: Failed to load Ephemeral cluster config file"
	// ContainerValidationEphemeralConfigError ...
	ContainerValidationEphemeralConfigError = "Container error: Ephemeral cluster configuration file validation failed"
	// ContainerSetEnvVarsError ...
	ContainerSetEnvVarsError = "Container error: Failed to set environment variables"
	// ContainerUnknownCommandError ...
	ContainerUnknownCommandError = "Container error: Container unknown command"
	// ContainerCreationEphemeralFailedError ...
	ContainerCreationEphemeralFailedError = "Container error: Creation of Ephemeral cluster failed"
	// ContainerDeletionEphemeralFailedError ...
	ContainerDeletionEphemeralFailedError = "Container error: Deletion of Ephemeral cluster failed"
	// ContainerHelpCommandFailedError ...
	ContainerHelpCommandFailedError = "Container error: Help command failed"
	// ContainerUnknownError ...
	ContainerUnknownError = "Container error: Unknown error code"
	// NumberOfRetriesExceededError ...
	NumberOfRetriesExceededError = "number of retries exceeded"
	// MissingConfigError ...
	MissingConfigError = "Missing Ephemeral Cluster ConfigFilename"
	// MissingVolumeError ...
	MissingVolumeError = "Must specify volume bind for Bootstrap builder container"
	// VolumeMalFormedError ...
	VolumeMalFormedError = "Bootstrap Container volume mount is mal formed"
	// MissingContainerImageError ...
	MissingContainerImageError = "Missing Bootstrap Container image"
	// MissingContainerRuntimeError ...
	MissingContainerRuntimeError = "Missing Container runtime information"
	// InvalidBootstrapCommandError ...
	InvalidBootstrapCommandError = "Invalid Bootstrap Container command"
)

// ErrBootstrapContainerRun is returned when there is an error executing the container
type ErrBootstrapContainerRun struct {
	ExitCode int
	StdErr   string
	ErrMsg   string
}

func (e ErrBootstrapContainerRun) Error() string {
	return fmt.Sprintf("Error from boostrap container: %s exit Code: %d\nContainer Logs: %s",
		e.ErrMsg, e.ExitCode, e.StdErr)
}

// ErrNumberOfRetriesExceeded is returned when number of retries have exceeded a limit
type ErrNumberOfRetriesExceeded struct {
}

func (e ErrNumberOfRetriesExceeded) Error() string {
	return NumberOfRetriesExceededError
}

// ErrInvalidInput is returned when invalid values were passed to the Bootstrap Container
type ErrInvalidInput struct {
	What string
}

func (e ErrInvalidInput) Error() string {
	return fmt.Sprintf("Invalid Bootstrap Container input: %s", e.What)
}

// ErrVolumeMalFormed is returned when error volume mount is mal formed
type ErrVolumeMalFormed struct {
}

func (e ErrVolumeMalFormed) Error() string {
	return VolumeMalFormedError
}

// ErrInvalidBootstrapCommand is returned when invalid command is invoked
type ErrInvalidBootstrapCommand struct {
}

func (e ErrInvalidBootstrapCommand) Error() string {
	return InvalidBootstrapCommandError
}

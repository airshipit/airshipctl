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

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	// BootCmdCreate is the string for the command "create" Ephemeral cluster
	BootCmdCreate = "create"
	// BootCmdDelete is the string for the command "delete" Ephemeral cluster
	BootCmdDelete = "delete"
	// BootCmdHelp is the string for the command "help" for the Ephemeral cluster
	BootCmdHelp = "help"
	// BootVolumeSeparator is the string Container volume mount
	BootVolumeSeparator = ":"
	// BootNullString represents an empty string
	BootNullString = ""
	// BootHelpFilename is the help filename
	BootHelpFilename = "help.txt"
	// Bootstrap Environment Variables
	envBootstrapCommand = "BOOTSTRAP_COMMAND"
	envBootstrapConfig  = "BOOTSTRAP_CONFIG"
	envBootstrapVolume  = "BOOTSTRAP_VOLUME"
)

var exitCodeMap = map[int]string{
	1: ContainerLoadEphemeralConfigError,
	2: ContainerValidationEphemeralConfigError,
	3: ContainerSetEnvVarsError,
	4: ContainerUnknownCommandError,
	5: ContainerCreationEphemeralFailedError,
	6: ContainerDeletionEphemeralFailedError,
	7: ContainerHelpCommandFailedError,
	8: ContainerUnknownError,
}

// BootstrapContainerOptions structure used by the executor
type BootstrapContainerOptions struct {
	Container container.Container
	Cfg       *v1alpha1.BootConfiguration
	Sleep     func(d time.Duration)

	// optional fields for verbose output
	Debug bool
}

// VerifyInputs verify if all input data to the container is correct
func (options *BootstrapContainerOptions) VerifyInputs() error {
	if options.Cfg.BootstrapContainer.Volume == "" {
		return ErrInvalidInput{
			What: MissingVolumeError,
		}
	}

	if options.Cfg.BootstrapContainer.Image == "" {
		return ErrInvalidInput{
			What: MissingContainerImageError,
		}
	}

	if options.Cfg.BootstrapContainer.ContainerRuntime == "" {
		return ErrInvalidInput{
			What: MissingContainerRuntimeError,
		}
	}

	if options.Cfg.EphemeralCluster.ConfigFilename == "" {
		return ErrInvalidInput{
			What: MissingConfigError,
		}
	}

	vols := strings.Split(options.Cfg.BootstrapContainer.Volume, ":")
	switch {
	case len(vols) == 1:
		options.Cfg.BootstrapContainer.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		return ErrVolumeMalFormed{}
	}
	return nil
}

// GetContainerStatus returns the Bootstrap Container state
func (options *BootstrapContainerOptions) GetContainerStatus() (container.Status, error) {
	// Check status of the container, e.g., "running"
	state, err := options.Container.InspectContainer()
	if err != nil {
		return BootNullString, err
	}

	var exitCode int
	exitCode = state.ExitCode
	if exitCode > 0 {
		reader, err := options.Container.GetContainerLogs()
		if err != nil {
			log.Printf("Error while trying to retrieve the container logs")
			return BootNullString, err
		}

		containerError := ErrBootstrapContainerRun{}
		containerError.ExitCode = exitCode
		containerError.ErrMsg = exitCodeMap[exitCode]

		if reader != nil {
			logs := new(bytes.Buffer)
			_, err = logs.ReadFrom(reader)
			if err != nil {
				return BootNullString, err
			}
			reader.Close()
			containerError.StdErr = logs.String()
		}
		return state.Status, containerError
	}

	return state.Status, nil
}

// WaitUntilContainerExitsOrTimesout waits for the container to exit or time out
func (options *BootstrapContainerOptions) WaitUntilContainerExitsOrTimesout(
	maxRetries int,
	configFilename string,
	bootstrapCommand string) error {
	// Give 2 seconds before checking if container is still running
	// This period should be enough to detect some initial errors thrown by the container
	options.Sleep(2 * time.Second)

	// Wait until container finished executing bootstrap of ephemeral cluster
	status, err := options.GetContainerStatus()
	if err != nil {
		return err
	}
	if status == container.ExitedContainerStatus {
		// bootstrap container command execution completed
		return nil
	}
	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("Waiting for bootstrap container using %s config file to %s Ephemeral cluster (%d/%d)",
			configFilename, bootstrapCommand, attempt, maxRetries)
		// Wait for 15 seconds and check again bootstrap container state
		options.Sleep(15 * time.Second)
		status, err = options.GetContainerStatus()
		if err != nil {
			return err
		}
		if status == container.ExitedContainerStatus {
			// bootstrap container command execution completed
			return nil
		}
	}
	return ErrNumberOfRetriesExceeded{}
}

// CreateBootstrapContainer creates a Bootstrap Container
func (options *BootstrapContainerOptions) CreateBootstrapContainer() error {
	containerVolMount := options.Cfg.BootstrapContainer.Volume
	vols := []string{containerVolMount}
	log.Printf("Running default container command. Mounted dir: %s", vols)

	bootstrapCommand := options.Cfg.EphemeralCluster.BootstrapCommand
	configFilename := options.Cfg.EphemeralCluster.ConfigFilename
	envVars := []string{
		fmt.Sprintf("%s=%s", envBootstrapCommand, bootstrapCommand),
		fmt.Sprintf("%s=%s", envBootstrapConfig, configFilename),
		fmt.Sprintf("%s=%s", envBootstrapVolume, containerVolMount),
	}

	err := options.Container.RunCommand([]string{}, nil, vols, envVars)
	if err != nil {
		return err
	}

	maxRetries := 50
	switch bootstrapCommand {
	case BootCmdCreate:
		// Wait until container finished executing bootstrap of ephemeral cluster
		err = options.WaitUntilContainerExitsOrTimesout(maxRetries, configFilename, bootstrapCommand)
		if err != nil {
			log.Printf("Failed to create Ephemeral cluster using %s config file", configFilename)
			return err
		}
		log.Printf("Ephemeral cluster created successfully using %s config file", configFilename)
	case BootCmdDelete:
		// Wait until container finished executing bootstrap of ephemeral cluster
		err = options.WaitUntilContainerExitsOrTimesout(maxRetries, configFilename, bootstrapCommand)
		if err != nil {
			log.Printf("Failed to delete Ephemeral cluster using %s config file", configFilename)
			return err
		}
		log.Printf("Ephemeral cluster deleted successfully using %s config file", configFilename)
	case BootCmdHelp:
		// Display Ephemeral Config file format for help
		sepPos := strings.Index(containerVolMount, BootVolumeSeparator)
		helpPath := filepath.Join(containerVolMount[:sepPos], BootHelpFilename)

		// Display help.txt on stdout
		data, err := ioutil.ReadFile(helpPath)
		if err != nil {
			log.Printf("File reading %s error: %s", helpPath, err)
			return err
		}
		// Printing the help.txt content to stdout
		fmt.Println(string(data))

		// Delete help.txt file
		err = os.Remove(helpPath)
		if err != nil {
			log.Printf("Could not delete %s", helpPath)
			return err
		}
	default:
		return ErrInvalidBootstrapCommand{}
	}

	log.Printf("Ephemeral cluster %s command completed successfully.", bootstrapCommand)
	if !options.Debug {
		log.Print("Removing bootstrap container.")
		return options.Container.RmContainer()
	}

	return nil
}

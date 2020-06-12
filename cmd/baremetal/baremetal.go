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

package baremetal

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote"
)

const (
	flagLabel            = "labels"
	flagLabelShort       = "l"
	flagLabelDescription = "Label(s) to filter desired baremetal host documents"

	flagName            = "name"
	flagNameShort       = "n"
	flagNameDescription = "Name to filter desired baremetal host document"

	flagPhase            = "phase"
	flagPhaseDescription = "airshipctl phase that contains the desired baremetal host document(s)"
)

// NewBaremetalCommand creates a new command for interacting with baremetal using airshipctl.
func NewBaremetalCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	baremetalRootCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Perform actions on baremetal hosts",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(rootSettings.Debug, cmd.OutOrStderr())

			// Load or Initialize airship Config
			rootSettings.InitConfig()
		},
	}

	ejectMediaCmd := NewEjectMediaCommand(rootSettings)
	baremetalRootCmd.AddCommand(ejectMediaCmd)

	powerOffCmd := NewPowerOffCommand(rootSettings)
	baremetalRootCmd.AddCommand(powerOffCmd)

	powerOnCmd := NewPowerOnCommand(rootSettings)
	baremetalRootCmd.AddCommand(powerOnCmd)

	powerStatusCmd := NewPowerStatusCommand(rootSettings)
	baremetalRootCmd.AddCommand(powerStatusCmd)

	rebootCmd := NewRebootCommand(rootSettings)
	baremetalRootCmd.AddCommand(rebootCmd)

	remoteDirectCmd := NewRemoteDirectCommand(rootSettings)
	baremetalRootCmd.AddCommand(remoteDirectCmd)

	return baremetalRootCmd
}

// GetHostSelections builds a list of selectors that can be passed to a manager
// using the name and label flags passed to airshipctl.
func GetHostSelections(name string, labels string) []remote.HostSelector {
	var selectors []remote.HostSelector
	if name != "" {
		selectors = append(selectors, remote.ByName(name))
	}

	if labels != "" {
		selectors = append(selectors, remote.ByLabel(labels))
	}

	return selectors
}

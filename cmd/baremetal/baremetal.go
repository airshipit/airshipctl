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
)

const (
	flagPhase            = "phase"
	flagPhaseDescription = "airshipctl phase that contains the desired baremetal host document(s)"
)

// NewBaremetalCommand creates a new command for interacting with baremetal using airshipctl.
func NewBaremetalCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Perform actions on baremetal hosts",
	}

	isoGenCmd := NewISOGenCommand(rootSettings)
	cmd.AddCommand(isoGenCmd)

	powerOffCmd := NewPowerOffCommand(rootSettings)
	cmd.AddCommand(powerOffCmd)

	powerStatusCmd := NewPowerStatusCommand(rootSettings)
	cmd.AddCommand(powerStatusCmd)

	rebootCmd := NewRebootCommand(rootSettings)
	cmd.AddCommand(rebootCmd)

	remoteDirectCmd := NewRemoteDirectCommand(rootSettings)
	cmd.AddCommand(remoteDirectCmd)

	return cmd
}

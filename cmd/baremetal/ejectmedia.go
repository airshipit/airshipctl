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
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// NewEjectMediaCommand provides a command to eject media attached to a baremetal host.
func NewEjectMediaCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	var labels string
	var name string
	var phase string

	cmd := &cobra.Command{
		Use:   "ejectmedia",
		Short: "Eject media attached to a baremetal host",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := remote.NewManager(rootSettings, phase, remote.ByLabel(labels), remote.ByName(name))
			if err != nil {
				return err
			}

			for _, host := range m.Hosts {
				if err := host.EjectVirtualMedia(host.Context); err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "All media ejected from host %s\n", host.HostName)
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&labels, flagLabel, flagLabelShort, "", flagLabelDescription)
	flags.StringVarP(&name, flagName, flagNameShort, "", flagNameDescription)
	flags.StringVar(&phase, flagPhase, config.BootstrapPhase, flagPhaseDescription)

	return cmd
}

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// NewPowerOffCommand provides a command to shutdown a remote host.
func NewPowerOffCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	powerOffCmd := &cobra.Command{
		Use:   "poweroff SYSTEM_ID",
		Short: "Shutdown a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := remote.NewAdapter(rootSettings)
			if err != nil {
				return err
			}

			if err := a.OOBClient.SystemPowerOff(a.Context, args[0]); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Remote host %s powered off\n", args[0])

			return nil
		},
	}

	return powerOffCmd
}

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
	"opendev.org/airship/airshipctl/pkg/inventory"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
)

var (
	powerOffCommand = "poweroff"

	powerOffLong = fmt.Sprintf(`
Power off baremetal hosts
%s
`, selectorsDescription)

	powerOffExample = fmt.Sprintf(bmhActionExampleTemplate, powerOffCommand)
)

// NewPowerOffCommand provides a command to shutdown a remote host.
func NewPowerOffCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     powerOffCommand,
		Short:   "Shutdown a baremetal hosts",
		Long:    powerOffLong[1:],
		Example: powerOffExample[1:],
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.BMHAction(ifc.BaremetalOperationPowerOff)
		},
	}

	initFlags(options, cmd)
	initAllFlag(options, cmd)

	return cmd
}

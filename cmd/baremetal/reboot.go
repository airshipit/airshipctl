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
	rebootCommand = "reboot"

	rebootLong = fmt.Sprintf(`
Reboot bare metal host(s). %s
`, selectorsDescription)

	rebootExample = fmt.Sprintf(bmhActionExampleTemplate, rebootCommand)
)

// NewRebootCommand provides a command with the capability to reboot baremetal hosts.
func NewRebootCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     rebootCommand,
		Long:    rebootLong[1:],
		Short:   "Airshipctl command to reboot host(s)",
		Example: rebootExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.BMHAction(ifc.BaremetalOperationReboot)
		},
	}

	initFlags(options, cmd)
	initAllFlag(options, cmd)

	return cmd
}

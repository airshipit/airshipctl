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
	powerOnCommand = "poweron"

	powerOnLong = fmt.Sprintf(`
Power on baremetal hosts
%s
`, selectorsDescription)

	powerOnExample = fmt.Sprintf(bmhActionExampleTempalte, powerOnCommand)
)

// NewPowerOnCommand provides a command with the capability to power on baremetal hosts.
func NewPowerOnCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     powerOnCommand,
		Short:   "Power on a hosts",
		Long:    powerOnLong[1:],
		Example: powerOnExample[1:],
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.BMHAction(ifc.BaremetalOperationPowerOn)
		},
	}

	initFlags(options, cmd)
	initAllFlag(options, cmd)

	return cmd
}

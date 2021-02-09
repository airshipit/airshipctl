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
	ejectMediaCommand = "ejectmedia"

	ejectMediaLong = fmt.Sprintf(`
Eject media attached to a baremetal hosts
%s
`, selectorsDescription)

	ejectMediaExample = fmt.Sprintf(bmhActionExampleTempalte, ejectMediaCommand)
)

// NewEjectMediaCommand provides a command to eject media attached to a baremetal host.
func NewEjectMediaCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     ejectMediaCommand,
		Short:   "Eject media attached to a baremetal hosts",
		Long:    ejectMediaLong[1:],
		Example: ejectMediaExample[1:],
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.BMHAction(ifc.BaremetalOperationEjectVirtualMedia)
		},
	}

	initFlags(options, cmd)
	initAllFlag(options, cmd)

	return cmd
}

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

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/inventory"
)

var (
	powerStatusLong = `
Retrieve the power status of a bare metal host. It targets a bare metal host from airship inventory
based on the --name, --namespace, --label and --timeout flags provided.
`

	powerStatusExample = `
To get power status of host with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal powerstatus --name rdm9r3s3

To get power status of host with name rdm9r3s3 in metal3 namespace
# airshipctl baremetal powerstatus --name rdm9r3s3 --namespace metal3

To get power status of host with a label 'foo=bar'
# airshipctl baremetal powerstatus --labels "foo=bar"
`
)

// NewPowerStatusCommand provides a command to retrieve the power status of a baremetal host.
func NewPowerStatusCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "powerstatus",
		Short:   "Airshipctl command to retrieve the power status of a bare metal host",
		Long:    powerStatusLong[1:],
		Example: powerStatusExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.PowerStatus(cmd.OutOrStdout())
		},
	}

	initFlags(options, cmd)

	return cmd
}

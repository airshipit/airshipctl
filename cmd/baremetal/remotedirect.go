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
	remoteDirectLong = `
Bootstrap bare metal host. It targets bare metal host from airship inventory based
on the --iso-url, --name, --namespace, --label and --timeout flags provided.
`

	remoteDirectExample = `
Perform action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal remotedirect --name rdm9r3s3

Perform action against hosts with name rdm9r3s3 in namespace metal3
# airshipctl baremetal remotedirect --name rdm9r3s3 --namespace metal3

Perform action against hosts with a label 'foo=bar'
# airshipctl baremetal remotedirect --labels "foo=bar"
`
)

// NewRemoteDirectCommand provides a command with the capability to perform remote direct operations.
func NewRemoteDirectCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remotedirect",
		Short:   "Airshipctl command to bootstrap the ephemeral host",
		Long:    remoteDirectLong[1:],
		Example: remoteDirectExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.RemoteDirect()
		},
	}
	initFlags(options, cmd)

	cmd.Flags().StringVar(&options.IsoURL, "iso-url", "", "specify iso url for host to boot from")

	return cmd
}

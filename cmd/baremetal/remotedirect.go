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

// NewRemoteDirectCommand provides a command with the capability to perform remote direct operations.
func NewRemoteDirectCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remotedirect",
		Short: "Bootstrap the ephemeral host",
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.RemoteDirect()
		},
	}
	initFlags(options, cmd)

	cmd.Flags().StringVar(&options.IsoURL, "iso-url", "", "specify iso url for host to boot from")

	return cmd
}

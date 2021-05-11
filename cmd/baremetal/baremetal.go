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
	"time"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/inventory"
)

// Action type is used to perform specific baremetal action
type Action int

const (
	baremetalLong = `
Provides commands that can be performed on bare metal host(s). The commands helps in
performing the basic boot and power operations on the bare metal host(s).
`
	flagLabel            = "labels"
	flagLabelShort       = "l"
	flagLabelDescription = "label(s) to filter desired bare metal host documents"

	flagName            = "name"
	flagNameDescription = "name to filter desired bare metal host document"

	flagNamespace            = "namespace"
	flagNamespaceSort        = "n"
	flagNamespaceDescription = "airshipctl phase that contains the desired bare metal host document(s)"

	flagTimeout            = "timeout"
	flagTimeoutDescription = "timeout on bare metal action"

	flagAll            = "all"
	flagAllDescription = "specify this to target all hosts in the inventory"
)

var (
	selectorsDescription = fmt.Sprintf(`The command will target bare metal hosts from airship inventory based on the
--%s, --%s and --%s flags provided. If no flags are provided, airshipctl will select all bare metal hosts in the
inventory.`, flagName, flagNamespace, flagLabel)

	bmhActionExampleTemplate = `
Perform action against hosts with name rdm9r3s3 in all namespaces where the host is found
# airshipctl baremetal %[1]s --name rdm9r3s3

Perform action against hosts with name rdm9r3s3 in namespace metal3
# airshipctl baremetal %[1]s --name rdm9r3s3 --namespace metal3

Perform action against all hosts defined in inventory
# airshipctl baremetal %[1]s --all

Perform action against hosts with a label 'foo=bar'
# airshipctl baremetal %[1]s --labels "foo=bar"
`
)

// NewBaremetalCommand creates a new command for interacting with baremetal using airshipctl.
func NewBaremetalCommand(cfgFactory config.Factory) *cobra.Command {
	options := inventory.NewOptions(inventory.NewInventory(cfgFactory))
	baremetalRootCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Airshipctl command to manage bare metal host(s)",
		Long:  baremetalLong,
	}

	baremetalRootCmd.AddCommand(NewEjectMediaCommand(cfgFactory, options))
	baremetalRootCmd.AddCommand(NewPowerOffCommand(cfgFactory, options))
	baremetalRootCmd.AddCommand(NewPowerOnCommand(cfgFactory, options))
	baremetalRootCmd.AddCommand(NewPowerStatusCommand(cfgFactory, options))
	baremetalRootCmd.AddCommand(NewRebootCommand(cfgFactory, options))
	baremetalRootCmd.AddCommand(NewRemoteDirectCommand(cfgFactory, options))

	return baremetalRootCmd
}

func initFlags(options *inventory.CommandOptions, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&options.Labels, flagLabel, flagLabelShort, "", flagLabelDescription)
	flags.StringVar(&options.Name, flagName, "", flagNameDescription)
	flags.StringVarP(&options.Namespace, flagNamespace, flagNamespaceSort, "", flagNamespaceDescription)
	flags.DurationVar(&options.Timeout, flagTimeout, 10*time.Minute, flagTimeoutDescription)
}

func initAllFlag(options *inventory.CommandOptions, cmd *cobra.Command) {
	cmd.Flags().BoolVar(&options.All, flagAll, false, flagAllDescription)
}

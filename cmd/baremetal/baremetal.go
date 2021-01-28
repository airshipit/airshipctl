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
	"time"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/inventory"
)

// Action type is used to perform specific baremetal action
type Action int

const (
	flagLabel            = "labels"
	flagLabelShort       = "l"
	flagLabelDescription = "Label(s) to filter desired baremetal host documents"

	flagName            = "name"
	flagNameDescription = "Name to filter desired baremetal host document"

	flagNamespace            = "namespace"
	flagNamespaceSort        = "n"
	flagNamespaceDescription = "airshipctl phase that contains the desired baremetal host document(s)"

	flagTimeout            = "timeout"
	flagTimeoutDescription = "timeout on baremetal action"
)

// NewBaremetalCommand creates a new command for interacting with baremetal using airshipctl.
func NewBaremetalCommand(cfgFactory config.Factory) *cobra.Command {
	options := inventory.NewOptions(inventory.NewInventory(cfgFactory))
	baremetalRootCmd := &cobra.Command{
		Use:   "baremetal",
		Short: "Perform actions on baremetal hosts",
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

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
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// Action type is used to perform specific baremetal action
type Action int

const (
	flagLabel            = "labels"
	flagLabelShort       = "l"
	flagLabelDescription = "Label(s) to filter desired baremetal host documents"

	flagName            = "name"
	flagNameShort       = "n"
	flagNameDescription = "Name to filter desired baremetal host document"

	flagPhase            = "phase"
	flagPhaseDescription = "airshipctl phase that contains the desired baremetal host document(s)"

	ejectAction Action = iota
	powerOffAction
	powerOnAction
	powerStatusAction
	rebootAction
	remoteDirectAction
)

// CommonOptions is used to store common variables from cmd flags for baremetal command group
type CommonOptions struct {
	labels string
	name   string
	phase  string
}

// NewBaremetalCommand creates a new command for interacting with baremetal using airshipctl.
func NewBaremetalCommand(cfgFactory config.Factory) *cobra.Command {
	options := &CommonOptions{}
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

func initFlags(options *CommonOptions, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&options.labels, flagLabel, flagLabelShort, "", flagLabelDescription)
	flags.StringVarP(&options.name, flagName, flagNameShort, "", flagNameDescription)
	flags.StringVar(&options.phase, flagPhase, config.BootstrapPhase, flagPhaseDescription)
}

func performAction(cfgFactory config.Factory, options *CommonOptions, action Action, writer io.Writer) error {
	cfg, err := cfgFactory()
	if err != nil {
		return err
	}

	selectors := GetHostSelections(options.name, options.labels)
	m, err := remote.NewManager(cfg, options.phase, selectors...)
	if err != nil {
		return err
	}

	return selectAction(m, cfg, action, writer)
}

func selectAction(m *remote.Manager, cfg *config.Config, action Action, writer io.Writer) error {
	if action == remoteDirectAction {
		if len(m.Hosts) != 1 {
			return remote.NewRemoteDirectErrorf("more than one node defined as the ephemeral node")
		}

		ephemeralHost := m.Hosts[0]
		return ephemeralHost.DoRemoteDirect(cfg)
	}

	ctx := context.Background()
	for _, host := range m.Hosts {
		switch action {
		case ejectAction:
			if err := host.EjectVirtualMedia(ctx); err != nil {
				return err
			}

			fmt.Fprintf(writer, "All media ejected from host '%s'.\n", host.HostName)
		case powerOffAction:
			if err := host.SystemPowerOff(ctx); err != nil {
				return err
			}

			fmt.Fprintf(writer, "Powered off host '%s'.\n", host.HostName)
		case powerOnAction:
			if err := host.SystemPowerOn(ctx); err != nil {
				return err
			}

			fmt.Fprintf(writer, "Powered on host '%s'.\n", host.HostName)
		case powerStatusAction:
			powerStatus, err := host.SystemPowerStatus(ctx)
			if err != nil {
				return err
			}

			fmt.Fprintf(writer, "Host '%s' has power status: '%s'\n",
				host.HostName, powerStatus)
		case rebootAction:
			if err := host.RebootSystem(ctx); err != nil {
				return err
			}

			fmt.Fprintf(writer, "Rebooted host '%s'.\n", host.HostName)
		}
	}
	return nil
}

// GetHostSelections builds a list of selectors that can be passed to a manager
// using the name and label flags passed to airshipctl.
func GetHostSelections(name string, labels string) []remote.HostSelector {
	var selectors []remote.HostSelector
	if name != "" {
		selectors = append(selectors, remote.ByName(name))
	}

	if labels != "" {
		selectors = append(selectors, remote.ByLabel(labels))
	}

	if len(selectors) == 0 {
		selectors = append(selectors, remote.All())
	}

	return selectors
}

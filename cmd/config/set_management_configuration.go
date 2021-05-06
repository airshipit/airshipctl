/*
  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	setManagementConfigLong = `
Creates or modifies management config information based on the MGMT_CONFIG_NAME passed. The allowed set
of optional flags are management-type, system-action-retries and system-reboot-delay. Use --use-proxy
and --insecure to enable proxy and insecure options respectively.
`

	setManagementConfigExample = `
Create management configuration
# airshipctl config set-management-config default

Create or update management configuration named "default" with retry and to enable insecure options
# airshipctl config set-management-config default --insecure --system-action-retries 40

Enable proxy for "test" management configuration
# airshipctl config set-management-config test --use-proxy
`
	flagInsecure            = "insecure"
	flagInsecureDescription = "ignore SSL certificate verification on out-of-band management requests"

	flagManagementType            = "management-type"
	flagManagementTypeDescription = "set the out-of-band management type"

	flagUseProxy            = "use-proxy"
	flagUseProxyDescription = "use the proxy configuration specified in the local environment"

	flagSystemActionRetries            = "system-action-retries"
	flagSystemActionRetriesDescription = "set the number of attempts to poll a host for a status"

	flagSystemRebootDelay            = "system-reboot-delay"
	flagSystemRebootDelayDescription = "set the number of seconds to wait between power actions (e.g. shutdown, startup)"
)

// NewSetManagementConfigCommand creates a command for creating and modifying clusters
// in the airshipctl config file.
func NewSetManagementConfigCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ManagementConfiguration{}
	cmd := &cobra.Command{
		Use:     "set-management-config MGMT_CONFIG_NAME",
		Short:   "Airshipctl command to create/modify out-of-band management configuration in airshipctl config file",
		Long:    setManagementConfigLong,
		Example: setManagementConfigExample,
		Args:    cobra.ExactArgs(1),
		RunE:    setManagementConfigRunE(cfgFactory, o),
	}

	addSetManagementConfigFlags(cmd, o)
	return cmd
}

func addSetManagementConfigFlags(cmd *cobra.Command, o *config.ManagementConfiguration) {
	flags := cmd.Flags()

	flags.BoolVar(&o.Insecure, flagInsecure, false, flagInsecureDescription)
	flags.StringVar(&o.Type, flagManagementType, redfish.ClientType, flagManagementTypeDescription)
	flags.BoolVar(&o.UseProxy, flagUseProxy, true, flagUseProxyDescription)
	flags.IntVar(&o.SystemActionRetries, flagSystemActionRetries,
		config.DefaultSystemActionRetries, flagSystemActionRetriesDescription)
	flags.IntVar(&o.SystemRebootDelay, flagSystemRebootDelay,
		config.DefaultSystemRebootDelay, flagSystemRebootDelayDescription)
}

func setManagementConfigRunE(cfgFactory config.Factory, o *config.ManagementConfiguration) func(
	cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Go through all the flags that have been set
		var opts []config.ManagementConfigOption
		fn := func(flag *pflag.Flag) {
			switch flag.Name {
			case flagInsecure:
				opts = append(opts, config.SetManagementConfigInsecure(o.Insecure))
			case flagManagementType:
				opts = append(opts, config.SetManagementConfigMgmtType(o.Type))
			case flagUseProxy:
				opts = append(opts, config.SetManagementConfigUseProxy(o.UseProxy))
			case flagSystemActionRetries:
				opts = append(opts, config.SetManagementConfigSystemActionRetries(o.SystemActionRetries))
			case flagSystemRebootDelay:
				opts = append(opts, config.SetManagementConfigSystemRebootDelay(o.SystemRebootDelay))
			}
		}
		cmd.Flags().Visit(fn)

		options := &config.RunSetManagementConfigOptions{
			CfgFactory:  cfgFactory,
			MgmtCfgName: args[0],
			Writer:      cmd.OutOrStdout(),
		}
		return options.RunSetManagementConfig(opts...)
	}
}

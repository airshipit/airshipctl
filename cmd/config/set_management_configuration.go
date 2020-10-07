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
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	flagInsecure            = "insecure"
	flagInsecureDescription = "Ignore SSL certificate verification on out-of-band management requests"

	flagManagementType            = "management-type"
	flagManagementTypeDescription = "Set the out-of-band management type"

	flagUseProxy            = "use-proxy"
	flagUseProxyDescription = "Use the proxy configuration specified in the local environment"
)

// NewSetManagementConfigCommand creates a command for creating and modifying clusters
// in the airshipctl config file.
func NewSetManagementConfigCommand(cfgFactory config.Factory) *cobra.Command {
	var insecure bool
	var managementType string
	var useProxy bool

	cmd := &cobra.Command{
		Use:   "set-management-config NAME",
		Short: "Modify an out-of-band management configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			name := args[0]
			managementCfg, err := cfg.GetManagementConfiguration(name)
			if err != nil {
				return err
			}

			var modified bool
			if cmd.Flags().Changed(flagInsecure) && insecure != managementCfg.Insecure {
				modified = true
				managementCfg.Insecure = insecure

				fmt.Fprintf(cmd.OutOrStdout(),
					"Option 'insecure' set to '%t' for management configuration '%s'.\n",
					managementCfg.Insecure, name)
			}

			if cmd.Flags().Changed(flagManagementType) && managementType != managementCfg.Type {
				modified = true
				if err = managementCfg.SetType(managementType); err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(),
					"Option 'type' set to '%s' for management configuration '%s'.\n",
					managementCfg.Type, name)
			}

			if cmd.Flags().Changed(flagUseProxy) && useProxy != managementCfg.UseProxy {
				modified = true
				managementCfg.UseProxy = useProxy

				fmt.Fprintf(cmd.OutOrStdout(),
					"Option 'useproxy' set to '%t' for management configuration '%s'\n",
					managementCfg.UseProxy, name)
			}

			if !modified {
				fmt.Fprintf(cmd.OutOrStdout(),
					"Management configuration '%s' not modified. No new settings.\n", name)
				return nil
			}

			return cfg.PersistConfig(true)
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(&insecure, flagInsecure, false, flagInsecureDescription)
	flags.StringVar(&managementType, flagManagementType, redfish.ClientType, flagManagementTypeDescription)
	flags.BoolVar(&useProxy, flagUseProxy, true, flagUseProxyDescription)

	return cmd
}

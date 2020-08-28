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
	"sort"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
)

const getManagementConfigExample = `
# View all defined management configurations
airshipctl config get-management-configs

# View a specific management configuration named "default"
airshipctl config get-management-config default
`

// NewGetManagementConfigCommand creates a command that enables printing a management configuration to stdout.
func NewGetManagementConfigCommand(cfgFactory config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-management-config [NAME]",
		Short:   "View a management config or all management configs defined in the airshipctl config",
		Example: getManagementConfigExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"get-management-configs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				name := args[0]

				config, err := cfg.GetManagementConfiguration(name)
				if err != nil {
					return err
				}

				fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n%s\n", name, config.String())

				return nil
			}

			if len(cfg.ManagementConfiguration) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No management configurations defined.")

				return nil
			}

			// Print all of the management configurations in order by name
			keys := make([]string, 0, len(cfg.ManagementConfiguration))
			for key := range cfg.ManagementConfiguration {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				config := cfg.ManagementConfiguration[key]
				fmt.Fprintf(cmd.OutOrStdout(), "name: %s\n%s\n", key, config.String())
			}

			return nil
		},
	}

	return cmd
}

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

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	getEncryptionConfigsLong = `
Display a specific encryption config information, or all defined encryption configs if no name is provided.
`

	getEncryptionConfigsExample = `
# List all the encryption configs airshipctl knows about
airshipctl config get-encryption-configs

# Display a specific encryption config
airshipctl config get-encryption-config exampleConfig
`
)

// NewGetEncryptionConfigCommand creates a command that enables printing an encryption configuration to stdout.
func NewGetEncryptionConfigCommand(cfgFactory config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-encryption-config NAME",
		Short:   "Get an encryption config information from the airshipctl config",
		Long:    getEncryptionConfigsLong[1:],
		Example: getEncryptionConfigsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"get-encryption-configs"},
		RunE: func(cmd *cobra.Command, args []string) error {
			airconfig, err := cfgFactory()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				name := args[0]
				encryptionConfig, exists := airconfig.EncryptionConfigs[name]
				if !exists {
					return config.ErrEncryptionConfigurationNotFound{
						Name: fmt.Sprintf("Encryption Config with name '%s'", name),
					}
				}
				fmt.Fprintln(cmd.OutOrStdout(), encryptionConfig)
			} else {
				encryptionConfigs := airconfig.GetEncryptionConfigs()
				if len(encryptionConfigs) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No Encryption Config found in the configuration.")
				}
				for _, encryptionConfig := range encryptionConfigs {
					fmt.Fprintln(cmd.OutOrStdout(), encryptionConfig)
				}
			}
			return nil
		},
	}

	return cmd
}

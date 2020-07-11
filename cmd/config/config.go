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
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// NewConfigCommand creates a command for interacting with the airshipctl configuration.
func NewConfigCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	configRootCmd := &cobra.Command{
		Use:                   "config",
		DisableFlagsInUseLine: true,
		Short:                 "Manage the airshipctl config file",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(rootSettings.Debug, cmd.OutOrStderr())
			if cmd.Use == "init" {
				rootSettings.Create = true
			}
			// Load or Initialize airship Config
			rootSettings.InitConfig()
		},
	}

	configRootCmd.AddCommand(NewGetAuthInfoCommand(rootSettings))
	configRootCmd.AddCommand(NewSetAuthInfoCommand(rootSettings))

	configRootCmd.AddCommand(NewGetClusterCommand(rootSettings))
	configRootCmd.AddCommand(NewSetClusterCommand(rootSettings))

	configRootCmd.AddCommand(NewGetContextCommand(rootSettings))
	configRootCmd.AddCommand(NewSetContextCommand(rootSettings))

	configRootCmd.AddCommand(NewGetManagementConfigCommand(rootSettings))
	configRootCmd.AddCommand(NewSetManagementConfigCommand(rootSettings))

	configRootCmd.AddCommand(NewImportCommand(rootSettings))
	configRootCmd.AddCommand(NewInitCommand(rootSettings))
	configRootCmd.AddCommand(NewUseContextCommand(rootSettings))

	configRootCmd.AddCommand(NewGetManifestCommand(rootSettings))
	configRootCmd.AddCommand(NewSetManifestCommand(rootSettings))

	return configRootCmd
}

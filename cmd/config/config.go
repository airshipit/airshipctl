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
)

// NewConfigCommand creates a command object for the airshipctl "config" , and adds all child commands to it.
func NewConfigCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	configRootCmd := &cobra.Command{
		Use:                   "config",
		DisableFlagsInUseLine: true,
		Short:                 "Modify airshipctl config files",
		Long: `Modify airshipctl config files using subcommands
like "airshipctl config set-context my-context" `,
	}
	configRootCmd.AddCommand(NewCmdConfigSetCluster(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigGetCluster(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigSetContext(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigGetContext(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigInit(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigSetAuthInfo(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigGetAuthInfo(rootSettings))
	configRootCmd.AddCommand(NewCmdConfigUseContext(rootSettings))

	return configRootCmd
}

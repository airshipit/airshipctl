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

package document

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/document/pull"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewPullCommand creates a new command for pulling airship document repositories
// initConfig determines whether it's appropriate to load configuration from
// disk; e.g. this is skipped when unit testing the command.
func NewPullCommand(rootSettings *environment.AirshipCTLSettings, initConfig bool) *cobra.Command {
	settings := pull.Settings{AirshipCTLSettings: rootSettings}
	documentPullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pulls documents from remote git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			if initConfig {
				// Load or Initialize airship Config
				rootSettings.InitConfig()
			}
			return settings.Pull()
		},
	}

	return documentPullCmd
}

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

package image

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// NewImageCommand creates a new command for managing ISO images using airshipctl.
func NewImageCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	imageRootCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage ISO image creation",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(rootSettings.Debug, cmd.OutOrStderr())

			// Load or Initialize airship Config
			rootSettings.InitConfig()
		},
	}

	imageBuildCmd := NewImageBuildCommand(rootSettings)
	imageRootCmd.AddCommand(imageBuildCmd)

	return imageRootCmd
}

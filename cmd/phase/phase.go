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

package phase

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	clusterLong = `
This command provides capabilities for interacting with phases,
such as getting list and applying specific one.
`
)

// NewPhaseCommand creates a command for interacting with phases
func NewPhaseCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	phaseRootCmd := &cobra.Command{
		Use:   "phase",
		Short: "Manage phases",
		Long:  clusterLong[1:],
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(rootSettings.Debug, cmd.OutOrStderr())

			// Load or Initialize airship Config
			rootSettings.InitConfig()
		},
	}

	phaseRootCmd.AddCommand(NewApplyCommand(rootSettings, client.DefaultClient))
	phaseRootCmd.AddCommand(NewRenderCommand(rootSettings))
	phaseRootCmd.AddCommand(NewPlanCommand(rootSettings))

	return phaseRootCmd
}

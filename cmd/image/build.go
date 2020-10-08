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

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

// NewImageBuildCommand creates a new command with the capability to build an ISO image.
func NewImageBuildCommand(cfgFactory config.Factory) *cobra.Command {
	var progress bool
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build ISO image",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := &phase.RunCommand{
				Factory: cfgFactory,
				Options: phase.RunFlags{Progress: progress},
			}
			p.Options.PhaseID.Name = config.BootstrapPhase
			return p.RunE()
		},
	}

	flags := cmd.Flags()
	flags.BoolVar(
		&progress,
		"progress",
		false,
		"show progress")

	return cmd
}

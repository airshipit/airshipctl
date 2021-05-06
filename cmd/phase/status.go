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

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	statusLong = `
Get the status of a phase such as ephemeral-control-plane, target-initinfra etc...
To list the phases associated with a site, run 'airshipctl phase list'.
`
	statusExample = `
Status of initinfra phase
# airshipctl phase status ephemeral-control-plane
`
)

// NewStatusCommand creates a command to find status of specific phase
func NewStatusCommand(cfgFactory config.Factory) *cobra.Command {
	ph := &phase.StatusCommand{
		Factory: cfgFactory,
		Options: phase.StatusFlags{},
	}

	statusCmd := &cobra.Command{
		Use:     "status PHASE_NAME",
		Short:   "Airshipctl command to show status of the phase",
		Long:    statusLong,
		Args:    cobra.ExactArgs(1),
		Example: statusExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			ph.Options.PhaseID.Name = args[0]
			return ph.RunE()
		},
	}
	return statusCmd
}

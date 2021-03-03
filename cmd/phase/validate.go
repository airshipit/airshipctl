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
	validLong = `Command which would validate that the phase contains ` +
		`the required documents to run the phase.
`

	validExample = `
# validate initinfra phase
airshipctl phase validate initinfra
`
)

// NewValidateCommand creates a command to assert that a phase is valid is to actually run the phase.
func NewValidateCommand(cfgFactory config.Factory) *cobra.Command {
	p := &phase.ValidateCommand{
		Options: phase.ValidateFlags{},
		Factory: cfgFactory,
	}
	validCmd := &cobra.Command{
		Use:     "validate PHASE_NAME",
		Short:   "Assert that a phase is valid",
		Long:    validLong,
		Args:    cobra.ExactArgs(1),
		Example: validExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Options.PhaseID.Name = args[0]
			return p.RunE()
		},
	}

	return validCmd
}

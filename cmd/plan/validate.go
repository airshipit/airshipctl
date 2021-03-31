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

package plan

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	validateLong = `
Run life-cycle phase validation which was defined in document model.
`
)

// NewValidateCommand creates a command which performs validation of particular phase plan
func NewValidateCommand(cfgFactory config.Factory) *cobra.Command {
	r := &phase.PlanValidateCommand{
		Factory: cfgFactory,
		Options: phase.PlanValidateFlags{},
	}
	runCmd := &cobra.Command{
		Use:   "validate PLAN_NAME",
		Short: "Validate plan",
		Long:  validateLong[1:],
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r.Options.PlanID.Name = args[0]
			return r.RunE()
		},
	}

	return runCmd
}

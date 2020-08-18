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
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	// TODO (kkalynovskyi) when different phase executors will be implmeneted, and their description is more clear,
	// add documentation here. also consider adding dynamic phase descriptions based on executors.
	// TODO (kkalynovskyi) when this command is fully functional and phase executors are developed
	// remove phase apply command
	runLong    = `Run specific life-cycle phase such as ephemeral-control-plane, target-initinfra etc...`
	runExample = `
# Run initinfra phase
airshipctl phase run ephemeral-control-plane
`
)

// NewRunCommand creates a command to run specific phase
func NewRunCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	p := &phase.Cmd{
		AirshipCTLSettings: rootSettings,
		Processor:          events.NewDefaultProcessor(utils.Streams()),
	}

	runCmd := &cobra.Command{
		Use:     "run PHASE_NAME",
		Short:   "Run phase",
		Long:    runLong,
		Args:    cobra.ExactArgs(1),
		Example: runExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Exec(args[0])
		},
	}
	flags := runCmd.Flags()
	flags.BoolVar(
		&p.DryRun,
		"dry-run",
		false,
		"simulate phase execution")
	return runCmd
}

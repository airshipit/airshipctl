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
	"github.com/spf13/pflag"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	runLong = `
Run a plan defined in the site manifest. Specify the plan using the mandatory parameter PLAN_NAME.
To get list of plans associated for a site, run 'airshipctl plan list'.
`
	runExample = `
Run plan named iso
# airshipctl plan run iso

Perform a dry run of a plan
# airshipctl plan run iso --dry-run
`
)

// NewRunCommand creates a command which execute a particular phase plan
func NewRunCommand(cfgFactory config.Factory) *cobra.Command {
	r := &phase.PlanRunCommand{Factory: cfgFactory}
	f := &phase.PlanRunFlags{}

	runCmd := &cobra.Command{
		Use:     "run PLAN_NAME",
		Short:   "Airshipctl command to run plan",
		Long:    runLong[1:],
		Example: runExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r.PlanID.Name = args[0]
			fn := func(flag *pflag.Flag) {
				switch flag.Name {
				case "dry-run":
					r.Options.DryRun = f.DryRun
				case "wait-timeout":
					r.Options.Timeout = &f.Timeout
				case "resume-from":
					r.Options.ResumeFromPhase = f.ResumeFromPhase
				}
			}
			cmd.Flags().Visit(fn)
			return r.RunE()
		},
	}

	flags := runCmd.Flags()
	flags.StringVar(&f.ResumeFromPhase, "resume-from", "", "skip all phases before the specified one")
	flags.BoolVar(&f.DryRun, "dry-run", false, "simulate phase execution")
	flags.DurationVar(&f.Timeout, "wait-timeout", 0, "wait timeout")
	return runCmd
}

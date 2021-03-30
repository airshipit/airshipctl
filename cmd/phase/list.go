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
	cmdLong = `
List life-cycle phases which were defined in document model by group.
Phases within a group are executed sequentially. Multiple phase groups
are executed in parallel.
`
	listExample = `
# List phases of phasePlan
airshipctl phase list --plan phasePlan

# To output the contents to table (default operation)
airshipctl phase list --plan phasePlan -o table

# To output the contents to yaml
airshipctl phase list --plan phasePlan -o yaml

# List all phases
airshipctl phase list

# List phases with clustername
airshipctl phase list --cluster-name clustername
`
)

// NewListCommand creates a command which prints available phases
func NewListCommand(cfgFactory config.Factory) *cobra.Command {
	p := &phase.ListCommand{Factory: cfgFactory}

	planCmd := &cobra.Command{
		Use:     "list PHASE_NAME",
		Short:   "List phases",
		Long:    cmdLong[1:],
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Writer = cmd.OutOrStdout()
			return p.RunE()
		},
	}
	addListFlags(p, planCmd)
	return planCmd
}

// addListFlags adds flags for phase list sub-command
func addListFlags(options *phase.ListCommand, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(
		&options.ClusterName,
		"cluster-name",
		"c",
		"",
		"filter documents by cluster name")

	flags.StringVar(
		&options.PlanID.Name,
		"plan",
		"",
		"Plan name of a plan")

	flags.StringVarP(
		&options.OutputFormat,
		"output", "o", "table", "'table' "+
			"and 'yaml' are available "+
			"output formats")
}

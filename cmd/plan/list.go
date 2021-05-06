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
	listLong = `
List plans defined in site manifest.
`

	listExample = `
List plan
# airshipctl plan list

List plan(yaml output format)
# airshipctl plan list -o yaml

List plan(table output format)
# airshipctl plan list -o table`
)

// NewListCommand creates a command which prints available phase plans
func NewListCommand(cfgFactory config.Factory) *cobra.Command {
	p := &phase.PlanListCommand{Factory: cfgFactory}

	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Airshipctl command to list plans",
		Long:    listLong[1:],
		Example: listExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			p.Writer = cmd.OutOrStdout()
			return p.RunE()
		},
	}
	flags := listCmd.Flags()
	flags.StringVarP(&p.Options.FormatType, "output", "o", "table",
		"output format. Supported formats are 'table' and 'yaml'")
	return listCmd
}

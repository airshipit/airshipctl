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
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/util"
)

const (
	cmdLong = `
List life-cycle phases which were defined in document model by group.
Phases within a group are executed sequentially. Multiple phase groups
are executed in parallel.
`
)

// NewPlanCommand creates a command which prints available phases
func NewPlanCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	p := &phase.Cmd{AirshipCTLSettings: rootSettings}

	planCmd := &cobra.Command{
		Use:   "plan",
		Short: "List phases",
		Long:  cmdLong[1:],
		RunE: func(cmd *cobra.Command, args []string) error {
			phases, err := p.Plan()
			if err != nil {
				return err
			}
			tw := util.NewTabWriter(cmd.OutOrStdout())
			defer tw.Flush()
			fmt.Fprintf(tw, "GROUP\tPHASE\n")
			for group, phaseList := range phases {
				fmt.Fprintf(tw, "%s\t\n", group)
				for _, phase := range phaseList {
					fmt.Fprintf(tw, "\t%s\n", phase)
				}
			}
			return nil
		},
	}
	return planCmd
}

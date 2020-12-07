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
)

const (
	planLong = `
This command provides capabilities for interacting with plan objects,
responsible for execution phases in groups
`
)

// NewPlanCommand creates a command for interacting with phases
func NewPlanCommand(cfgFactory config.Factory) *cobra.Command {
	planRootCmd := &cobra.Command{
		Use:   "plan",
		Short: "Manage plans",
		Long:  planLong[1:],
	}

	planRootCmd.AddCommand(NewListCommand(cfgFactory))
	planRootCmd.AddCommand(NewRunCommand(cfgFactory))

	return planRootCmd
}

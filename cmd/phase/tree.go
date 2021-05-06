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
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	treeLong = `
Get tree view of the kustomize entrypoints of a phase.
`

	treeExample = `
yaml explorer of a phase with relative path
# airshipctl phase tree /manifests/site/test-site/ephemeral/initinfra

yaml explorer of a phase with phase name
# airshipctl phase tree initinfra-ephemeral
`
)

// NewTreeCommand creates a command to get summarized tree view of the kustomize entrypoints of a phase
func NewTreeCommand(cfgFactory config.Factory) *cobra.Command {
	treeCmd := &cobra.Command{
		Use:     "tree PHASE_NAME",
		Short:   "Airshipctl command to show tree view of kustomize entrypoints of phase",
		Long:    treeLong[1:],
		Args:    cobra.ExactArgs(1),
		Example: treeExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			writer := cmd.OutOrStdout()
			p := &phase.TreeCommand{
				Factory: cfgFactory,
				PhaseID: ifc.ID{},
				Writer:  writer,
			}
			p.Argument = args[0]
			return p.RunE()
		},
	}
	return treeCmd
}

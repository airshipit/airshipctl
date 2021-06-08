/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	getContextLong = `
Displays information about contexts such as associated manifests, users, and clusters. It would display a specific
context information, or all defined context information if no name is provided.
`

	getContextExample = `
List all contexts
# airshipctl config get-contexts

Display the current context
# airshipctl config get-context --current

Display a specific context
# airshipctl config get-context exampleContext
`
)

// NewGetContextCommand creates a command for viewing cluster information
// defined in the airshipctl config file.
func NewGetContextCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ContextOptions{}
	cmd := &cobra.Command{
		Use:     "get-context CONTEXT_NAME",
		Short:   "Airshipctl command to get context(s) information from the airshipctl config",
		Long:    getContextLong[1:],
		Example: getContextExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"get-contexts"},
		RunE: func(cmd *cobra.Command, args []string) error {
			airconfig, err := cfgFactory()
			if err != nil {
				return err
			}
			if len(args) == 1 {
				o.Name = args[0]
			}

			if len(airconfig.GetContexts()) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No Contexts found in the configuration.")
			} else {
				return o.Print(airconfig, cmd.OutOrStdout())
			}
			return nil
		},
	}

	addGetContextFlags(o, cmd)
	return cmd
}

func addGetContextFlags(o *config.ContextOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.BoolVar(&o.CurrentContext, "current", false, "get the current context")
	flags.StringVar(&o.Format, "format", "yaml",
		"supported output format `yaml` or `table`, default is `yaml`")
}

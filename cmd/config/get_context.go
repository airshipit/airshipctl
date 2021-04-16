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
Display information about contexts such as associated manifests, users, and clusters.
`

	getContextExample = `
# List all contexts
airshipctl config get-contexts

# Display the current context
airshipctl config get-context --current

# Display a specific context
airshipctl config get-context exampleContext
`
)

// NewGetContextCommand creates a command for viewing cluster information
// defined in the airshipctl config file.
func NewGetContextCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ContextOptions{}
	cmd := &cobra.Command{
		Use:     "get-context CONTEXT_NAME",
		Short:   "Get context information from the airshipctl config",
		Long:    getContextLong[1:],
		Example: getContextExample,
		// Adding a maximum args cap for documentation purpose
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
			if o.Name == "" && !o.CurrentContext {
				contexts := airconfig.GetContexts()
				if len(contexts) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No Contexts found in the configuration.")
				}
				for _, context := range contexts {
					fmt.Fprintln(cmd.OutOrStdout(), context.String())
				}
				return nil
			}

			if o.CurrentContext {
				o.Name = airconfig.CurrentContext
			}

			context, err := airconfig.GetContext(o.Name)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), context.String())
			return nil
		},
	}

	addGetContextFlags(o, cmd)
	return cmd
}

func addGetContextFlags(o *config.ContextOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.BoolVar(
		&o.CurrentContext,
		"current",
		false,
		"get the current context")
}

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
	useContextLong = `
Switch to a different context defined in the airshipctl config file.
This command doesn't change a context for the kubeconfig file.
`

	useContextExample = `
# Switch to a context named "exampleContext" in airshipctl config file
airshipctl config use-context exampleContext
`
)

// NewUseContextCommand creates a command for switching to a defined airshipctl context.
func NewUseContextCommand(cfgFactory config.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "use-context NAME",
		Short:   "Switch to a different context",
		Long:    useContextLong[1:],
		Example: useContextExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			contextName := args[0]
			err = config.RunUseContext(contextName, cfg)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", contextName)

			return nil
		},
	}

	return cmd
}

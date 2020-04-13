/*
Copyright 2016 The Kubernetes Authors.

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
	"opendev.org/airship/airshipctl/pkg/environment"
)

var (
	useContextLong = "Switch to a new context defined in the airshipctl config file."

	useContextExample = `
# Switch to a context named "e2e"
airshipctl config use-context e2e`
)

// NewCmdConfigUseContext creates a command object for the "use-context" action, which
// switches to a defined airshipctl context.
func NewCmdConfigUseContext(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "use-context NAME",
		Short:   "Switch to a different airshipctl context.",
		Long:    useContextLong,
		Example: useContextExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			contextName := args[0]
			err := config.RunUseContext(contextName, rootSettings.Config)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Switched to context %q.\n", contextName)

			return nil
		},
	}

	return cmd
}

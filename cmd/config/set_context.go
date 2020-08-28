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
)

const (
	setContextLong = `
Create or modify a context in the airshipctl config files.
`

	setContextExample = `
# Create a new context named "exampleContext"
airshipctl config set-context exampleContext \
  --namespace=kube-system \
  --manifest=exampleManifest \
  --user=exampleUser
  --cluster-type=target

# Update the manifest of the current-context
airshipctl config set-context \
  --current \
  --manifest=exampleManifest
`
)

// NewSetContextCommand creates a command for creating and modifying contexts
// in the airshipctl config
func NewSetContextCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ContextOptions{}
	cmd := &cobra.Command{
		Use:     "set-context NAME",
		Short:   "Manage contexts",
		Long:    setContextLong[1:],
		Example: setContextExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nFlags := cmd.Flags().NFlag()
			if len(args) == 1 {
				// context name is made optional with --current flag added
				o.Name = args[0]
			}
			if o.Name != "" && nFlags == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q not modified. No new options provided.\n", o.Name)
				return nil
			}

			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			modified, err := config.RunSetContext(o, cfg, true)

			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q modified.\n", o.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q created.\n", o.Name)
			}
			return nil
		},
	}

	addSetContextFlags(o, cmd)
	return cmd
}

func addSetContextFlags(o *config.ContextOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.Cluster,
		"cluster",
		"",
		"set the cluster for the specified context")

	flags.StringVar(
		&o.AuthInfo,
		"user",
		"",
		"set the user for the specified context")

	flags.StringVar(
		&o.Manifest,
		"manifest",
		"",
		"set the manifest for the specified context")

	flags.StringVar(
		&o.Namespace,
		"namespace",
		"",
		"set the namespace for the specified context")

	flags.StringVar(
		&o.ClusterType,
		"cluster-type",
		"",
		"set the cluster-type for the specified context")

	flags.BoolVar(
		&o.Current,
		"current",
		false,
		"update the current context")
}

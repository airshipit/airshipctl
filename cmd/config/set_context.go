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
	setContextLong = `
Sets a context entry in arshipctl config.
Specifying a name that already exists will merge new fields on top of existing values for those fields.`

	setContextExample = fmt.Sprintf(`
# Create a completely new e2e context entry
airshipctl config set-context e2e --%v=kube-system --%v=manifest --%v=auth-info --%v=%v

# Update the current-context to e2e
airshipctl config set-context e2e

# Update attributes of the current-context
airshipctl config set-context --%s --%v=manifest`,
		config.FlagNamespace,
		config.FlagManifest,
		config.FlagAuthInfoName,
		config.FlagClusterType,
		config.Target,
		config.FlagCurrent,
		config.FlagManifest)
)

// NewCmdConfigSetContext creates a command object for the "set-context" action, which
// creates and modifies contexts in the airshipctl config
func NewCmdConfigSetContext(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.ContextOptions{}

	cmd := &cobra.Command{
		Use:     "set-context NAME",
		Short:   "Switch to a new context or update context values in the airshipctl config",
		Long:    setContextLong,
		Example: setContextExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Name = args[0]
			nFlags := cmd.Flags().NFlag()
			if nFlags == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q not modified. No new options provided.\n", o.Name)
				return nil
			}

			if len(args) == 1 {
				//context name is made optional with --current flag added
				o.Name = args[0]
			}
			modified, err := config.RunSetContext(o, rootSettings.Config(), true)

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
		config.FlagClusterName,
		"",
		"sets the "+config.FlagClusterName+" for the specified context in the airshipctl config")

	flags.StringVar(
		&o.AuthInfo,
		config.FlagAuthInfoName,
		"",
		"sets the "+config.FlagAuthInfoName+" for the specified context in the airshipctl config")

	flags.StringVar(
		&o.Manifest,
		config.FlagManifest,
		"",
		"sets the "+config.FlagManifest+" for the specified context in the airshipctl config")

	flags.StringVar(
		&o.Namespace,
		config.FlagNamespace,
		"",
		"sets the "+config.FlagNamespace+" for the specified context in the airshipctl config")

	flags.StringVar(
		&o.ClusterType,
		config.FlagClusterType,
		"",
		"sets the "+config.FlagClusterType+" for the specified context in the airshipctl config")

	flags.BoolVar(
		&o.Current,
		config.FlagCurrent,
		false,
		"use current context from airshipctl config")
}

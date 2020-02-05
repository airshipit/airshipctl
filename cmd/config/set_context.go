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
	"errors"
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
airshipctl config set-context e2e --%v=true`,
		config.FlagNamespace,
		config.FlagManifest,
		config.FlagAuthInfoName,
		config.FlagClusterType,
		config.Target,
		config.FlagCurrentContext)
)

// NewCmdConfigSetContext creates a command object for the "set-context" action, which
// defines a new Context airshipctl config.
func NewCmdConfigSetContext(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	theContext := &config.ContextOptions{}

	setcontextcmd := &cobra.Command{
		Use:     "set-context NAME",
		Short:   "Sets a context entry or updates current-context in the airshipctl config",
		Long:    setContextLong,
		Example: setContextExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			theContext.Name = cmd.Flags().Args()[0]
			modified, err := runSetContext(theContext, rootSettings.Config())
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q modified.\n", theContext.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Context %q created.\n", theContext.Name)
			}
			return nil
		},
	}

	sctxInitFlags(theContext, setcontextcmd)
	return setcontextcmd
}

func sctxInitFlags(o *config.ContextOptions, setcontextcmd *cobra.Command) {
	setcontextcmd.Flags().BoolVar(&o.CurrentContext, config.FlagCurrentContext, false,
		config.FlagCurrentContext+" for the context entry in airshipctl config")

	setcontextcmd.Flags().StringVar(&o.Cluster, config.FlagClusterName, o.Cluster,
		config.FlagClusterName+" for the context entry in airshipctl config")

	setcontextcmd.Flags().StringVar(&o.AuthInfo, config.FlagAuthInfoName, o.AuthInfo,
		config.FlagAuthInfoName+" for the context entry in airshipctl config")

	setcontextcmd.Flags().StringVar(&o.Manifest, config.FlagManifest, o.Manifest,
		config.FlagManifest+" for the context entry in airshipctl config")

	setcontextcmd.Flags().StringVar(&o.Namespace, config.FlagNamespace, o.Namespace,
		config.FlagNamespace+" for the context entry in airshipctl config")

	setcontextcmd.Flags().StringVar(&o.ClusterType, config.FlagClusterType, "",
		config.FlagClusterType+" for the context entry in airshipctl config")
}

func runSetContext(o *config.ContextOptions, airconfig *config.Config) (bool, error) {
	contextWasModified := false
	err := o.Validate()
	if err != nil {
		return contextWasModified, err
	}

	contextIWant := o.Name
	context, err := airconfig.GetContext(contextIWant)
	if err != nil {
		var cerr config.ErrMissingConfig
		if !errors.As(err, &cerr) {
			// An error occurred, but it wasn't a "missing" config error.
			return contextWasModified, err
		}

		if o.CurrentContext {
			return contextWasModified, config.ErrMissingConfig{}
		}
		// context didn't exist, create it
		// ignoring the returned added context
		airconfig.AddContext(o)
	} else {
		// Found the desired Current Context
		// Lets update it and be done.
		if o.CurrentContext {
			airconfig.CurrentContext = o.Name
		} else {
			// Context exists, lets update
			airconfig.ModifyContext(context, o)
		}
		contextWasModified = true
	}
	// Update configuration file just in time persistence approach
	if err := airconfig.PersistConfig(); err != nil {
		// Error that it didnt persist the changes
		return contextWasModified, config.ErrConfigFailed{}
	}

	return contextWasModified, nil
}

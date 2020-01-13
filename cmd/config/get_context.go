/*
Copyright 2014 The Kubernetes Authors.

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
	"io"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

var (
	getContextLong = (`Display a specific context, the current-context or all defined contexts if no name is provided`)

	getContextExample = fmt.Sprintf(`# List all the contexts  airshipctl knows about
airshipctl config get-context

# Display the current context
airshipctl config get-context --%v

# Display a specific Context
airshipctl config get-context e2e`,
		config.FlagCurrentContext)
)

// A Context refers to a particular cluster, however it does not specify which of the cluster types
// it relates to. Getting explicit  information about a particular context will depend
// on the ClusterType flag.

// NewCmdConfigGetContext returns a Command instance for 'config -Context' sub command
func NewCmdConfigGetContext(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	theContext := &config.ContextOptions{}
	getcontextcmd := &cobra.Command{
		Use:     "get-context NAME",
		Short:   getContextLong,
		Example: getContextExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				theContext.Name = args[0]
			}
			return runGetContext(theContext, cmd.OutOrStdout(), rootSettings.Config())
		},
	}

	gctxInitFlags(theContext, getcontextcmd)

	return getcontextcmd
}

func gctxInitFlags(o *config.ContextOptions, getcontextcmd *cobra.Command) {
	getcontextcmd.Flags().BoolVar(&o.CurrentContext, config.FlagCurrentContext, false,
		config.FlagCurrentContext+" to retrieve the current context entry in airshipctl config")
}

// runGetContext performs the execution of 'config get-Context' sub command
func runGetContext(o *config.ContextOptions, out io.Writer, airconfig *config.Config) error {
	if o.Name == "" && !o.CurrentContext {
		return getContexts(out, airconfig)
	}
	return getContext(o, out, airconfig)
}

func getContext(o *config.ContextOptions, out io.Writer, airconfig *config.Config) error {
	cName := o.Name
	if o.CurrentContext {
		cName = airconfig.CurrentContext
	}
	context, err := airconfig.GetContext(cName)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", context.PrettyString())
	return nil
}

func getContexts(out io.Writer, airconfig *config.Config) error {
	contexts, err := airconfig.GetContexts()
	if err != nil {
		return err
	}
	if contexts == nil {
		fmt.Fprint(out, "No Contexts found in the configuration.\n")
	}
	for _, context := range contexts {
		fmt.Fprintf(out, "%s", context.PrettyString())
	}
	return nil
}

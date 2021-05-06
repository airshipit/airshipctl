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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	setContextLong = `
Creates or modifies context in the airshipctl config file based on the CONTEXT_NAME passed or for the current context
if --current flag is specified. It accepts optional flags which include manifest name and management-config name.
`

	setContextExample = `
To create a new context named "exampleContext"
# airshipctl config set-context exampleContext --manifest=exampleManifest

To update the manifest of the current-context
# airshipctl config set-context --current --manifest=exampleManifest
`

	setContextManifestFlag         = "manifest"
	setContextManagementConfigFlag = "management-config"
	setContextCurrentFlag          = "current"
)

// NewSetContextCommand creates a command for creating and modifying contexts
// in the airshipctl config
func NewSetContextCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ContextOptions{}
	cmd := &cobra.Command{
		Use:     "set-context CONTEXT_NAME",
		Short:   "Airshipctl command to create/modify context in airshipctl config file",
		Long:    setContextLong[1:],
		Example: setContextExample,
		Args:    cobra.MaximumNArgs(1),
		RunE:    setContextRunE(cfgFactory, o),
	}

	addSetContextFlags(cmd, o)
	return cmd
}

func addSetContextFlags(cmd *cobra.Command, o *config.ContextOptions) {
	flags := cmd.Flags()

	flags.StringVar(&o.Manifest, setContextManifestFlag, "",
		"set the manifest for the specified context")
	flags.StringVar(&o.ManagementConfiguration, setContextManagementConfigFlag, "",
		"set the management config for the specified context")
	flags.BoolVar(&o.Current, setContextCurrentFlag, false,
		"update the current context")
}

func setContextRunE(cfgFactory config.Factory, o *config.ContextOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctxName := ""
		if len(args) == 1 {
			ctxName = args[0]
		}

		// Go through all the flags that have been set
		var opts []config.ContextOption
		fn := func(flag *pflag.Flag) {
			switch flag.Name {
			case setContextManifestFlag:
				opts = append(opts, config.SetContextManifest(o.Manifest))
			case setContextManagementConfigFlag:
				opts = append(opts, config.SetContextManagementConfig(o.ManagementConfiguration))
			}
		}
		cmd.Flags().Visit(fn)

		options := &config.RunSetContextOptions{
			CfgFactory: cfgFactory,
			CtxName:    ctxName,
			Current:    o.Current,
			Writer:     cmd.OutOrStdout(),
		}
		return options.RunSetContext(opts...)
	}
}

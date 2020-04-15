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

package document

import (
	"io/ioutil"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/document/plugin"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	pluginLong = `
This command is meant to be used as a kustomize exec plugin.

The command reads the configuration file CONFIG passed as a first argument and
determines a particular plugin to execute. Additional arguments may be passed
to this command and can be used by the particular plugin.

CONFIG must be a structured kubernetes manifest (i.e. resource) and must have
'apiVersion' and 'kind' keys. If the appropriate plugin was not found, the
command returns an error.
`

	pluginExample = `
# Perform a replacement on a deployment. Prior to running this command,
# the file '/tmp/replacement.yaml' should be created as follows:
---
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image

# The replacement can then be performed. Output defaults to stdout.
airshipctl document plugin /tmp/replacement.yaml
`
)

// NewPluginCommand creates a new command which can act as kustomize
// exec plugin.
func NewPluginCommand(rootSetting *environment.AirshipCTLSettings) *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:     "plugin CONFIG [ARGS]",
		Short:   "Run as a kustomize exec plugin",
		Long:    pluginLong[1:],
		Example: pluginExample,
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			return plugin.ConfigureAndRun(rootSetting, cfg, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
	return pluginCmd
}

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

var longDescription = `Subcommand reads configuration file CONFIG passed as
a first argument and determines a particular plugin to execute. Additional
arguments may be passed to this sub-command abd can be used by the
particular plugin. CONFIG file must be structured as kubernetes
manifest (i.e. resource) and must have 'apiVersion' and 'kind' keys.

Example:
$ cat /tmp/generator.yaml
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

$ airshipctl document plugin /tmp/generator.yaml

subcommand will try to identify appropriate plugin using apiVersion and
kind keys (a.k.a group, version, kind) as an identifier. If appropriate
plugin was not found command returns an error.
`

// NewDocumentPluginCommand creates a new command which can act as kustomize
// exec plugin.
func NewDocumentPluginCommand(rootSetting *environment.AirshipCTLSettings) *cobra.Command {
	pluginCmd := &cobra.Command{
		Use:   "plugin CONFIG [ARGS]",
		Short: "used as kustomize exec plugin",
		Long:  longDescription,
		Args:  cobra.MinimumNArgs(1),
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

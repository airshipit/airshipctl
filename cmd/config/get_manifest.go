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

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	getManifestsLong = `
Display a specific manifest information, or all defined manifests if no name is provided.
`

	getManifestsExample = `
# List all the manifests airshipctl knows about
airshipctl config get-manifests

# Display a specific manifest
airshipctl config get-manifest e2e
`
)

// NewGetManifestCommand creates a command for viewing the manifest information
// defined in the airshipctl config file.
func NewGetManifestCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.ManifestOptions{}
	cmd := &cobra.Command{
		Use:     "get-manifest NAME",
		Short:   "Get a manifest information from the airshipctl config",
		Long:    getManifestsLong[1:],
		Example: getManifestsExample,
		Args:    cobra.MaximumNArgs(1),
		Aliases: []string{"get-manifests"},
		RunE: func(cmd *cobra.Command, args []string) error {
			airconfig := rootSettings.Config
			if len(args) == 1 {
				o.Name = args[0]
				manifest, exists := airconfig.Manifests[o.Name]
				if !exists {
					return config.ErrMissingConfig{
						What: fmt.Sprintf("Manifest with name '%s'", o.Name),
					}
				}
				fmt.Fprintln(cmd.OutOrStdout(), manifest)
			} else {
				manifests := airconfig.GetManifests()
				if len(manifests) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No Manifest found in the configuration.")
				}
				for _, manifest := range manifests {
					fmt.Fprintln(cmd.OutOrStdout(), manifest)
				}
			}
			return nil
		},
	}

	return cmd
}

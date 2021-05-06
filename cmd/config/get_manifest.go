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
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	getManifestsLong = `
Displays a specific manifest information, or all defined manifests if no name is provided. The information
includes the repository details related to site manifest along with the local targetPath for them.
`

	getManifestsExample = `
List all the manifests
# airshipctl config get-manifests

Display a specific manifest
# airshipctl config get-manifest e2e
`
)

// NewGetManifestCommand creates a command for viewing the manifest information
// defined in the airshipctl config file.
func NewGetManifestCommand(cfgFactory config.Factory) *cobra.Command {
	var manifestName string
	cmd := &cobra.Command{
		Use:     "get-manifest MANIFEST_NAME",
		Short:   "Airshipctl command to get a specific or all manifest(s) information from the airshipctl config",
		Long:    getManifestsLong[1:],
		Example: getManifestsExample,
		Args:    GetManifestNArgs(&manifestName),
		Aliases: []string{"get-manifests"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return config.RunGetManifest(cfgFactory, manifestName, cmd.OutOrStdout())
		},
	}

	return cmd
}

// GetManifestNArgs is used to process arguments for get-manifest cmd
func GetManifestNArgs(manifestName *string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == "get-manifests" {
			return cobra.ExactArgs(0)(cmd, args)
		}
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		*manifestName = args[0]
		return nil
	}
}

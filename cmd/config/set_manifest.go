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
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	setManifestsLong = `
Create or modify a manifests in the airshipctl config file.
`

	setManifestsExample = `
# Create a new manifest
airshipctl config set-manifest exampleManifest \
  --repo exampleRepo \
  --url https://github.com/site \
  --branch master \
  --phase \
  --sub-path exampleSubpath \
  --target-path exampleTargetpath

# Change the phase repo for manifest
airshipctl config set-manifest e2e \
  --repo exampleRepo \
  --phase

# Change the sub-path for manifest
airshipctl config set-manifest e2e \
  --sub-path treasuremap/manifests/e2e

# Change the target-path for manifest
airshipctl config set-manifest e2e \
  --target-path /tmp/e2e
`
)

// NewSetManifestCommand creates a command for creating and modifying manifests
// in the airshipctl config file.
func NewSetManifestCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.ManifestOptions{}
	cmd := &cobra.Command{
		Use:     "set-manifest NAME",
		Short:   "Manage manifests in airship config",
		Long:    setManifestsLong[1:],
		Example: setManifestsExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			o.Name = args[0]
			modified, err := config.RunSetManifest(o, cfg, true)
			// Check if URL flag is passed with empty value
			if cmd.Flags().Changed("url") && o.URL == "" {
				log.Fatal("Repository URL cannot be empty.")
			}
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Manifest %q modified.\n", o.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Manifest %q created.\n", o.Name)
			}
			return nil
		},
	}

	addSetManifestFlags(o, cmd)
	return cmd
}

func addSetManifestFlags(o *config.ManifestOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.RepoName,
		"repo",
		"",
		"the name of the repository to be associated with this manifest")

	flags.StringVar(
		&o.URL,
		"url",
		"",
		"the repository url to be associated with this manifest")

	flags.StringVar(
		&o.Branch,
		"branch",
		"",
		"the branch to be associated with repository in this manifest")

	flags.StringVar(
		&o.CommitHash,
		"commithash",
		"",
		"the commit hash to be associated with repository in this manifest")

	flags.StringVar(
		&o.Tag,
		"tag",
		"",
		"the tag to be associated with repository in this manifest")

	flags.BoolVar(
		&o.Force,
		"force",
		false,
		"if set, enable force checkout in repository with this manifest")

	flags.BoolVar(
		&o.IsPhase,
		"phase",
		false,
		"if set, enable this repository as phase repository to be used with this manifest")

	flags.StringVar(
		&o.SubPath,
		"sub-path",
		"",
		"the sub path to be set for this manifest")

	flags.StringVar(
		&o.TargetPath,
		"target-path",
		"",
		"the target path to be set for this manifest")

	flags.StringVar(
		&o.MetadataPath,
		"metadata-path",
		"",
		"the metadata path to be set for this manifest")
}

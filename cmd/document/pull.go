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
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document/pull"
)

const (
	long = `
The remote manifests repositories as well as the target path where
the repositories will be cloned are defined in the airship config file.

By default the airship config file is initialized with the
repository "https://opendev.org/airship/treasuremap" as a source of
manifests and with the manifests target path "%s".
`

	pullExample = `
Pull manifests from remote repos
# airshipctl document pull
For the below sample airship config file, it will pull from remote repository where URL mentioned
to the target location /home/airship with manifests->treasuremap->repositories->airshipctl->checkout
options branch, commitHash & tag mentioned in manifest section.
In the URL section, instead of a remote repository location we can also mention already checkout directory,
in this case we need not use document pull otherwise, any temporary changes will be overwritten.
>>>>>>Sample Config File<<<<<<<<<
cat ~/.airship/config
apiVersion: airshipit.org/v1alpha1
contexts:
  ephemeral-cluster:
    managementConfiguration: treasuremap_config
    manifest: treasuremap
  target-cluster:
    managementConfiguration: treasuremap_config
    manifest: treasuremap
currentContext: ephemeral-cluster
kind: Config
managementConfiguration:
  treasuremap_config:
    insecure: true
    systemActionRetries: 30
    systemRebootDelay: 30
    type: redfish
manifests:
  treasuremap:
    inventoryRepositoryName: primary
    metadataPath: manifests/site/eric-test-site/metadata.yaml
    phaseRepositoryName: primary
    repositories:
      airshipctl:
        checkout:
          branch: ""
          commitHash: f4cb1c44e0283c38a8bc1be5b8d71020b5d30dfb
          force: false
          localBranch: false
          tag: ""
        url: https://opendev.org/airship/airshipctl.git
      primary:
        checkout:
          branch: ""
          commitHash: 5556edbd386191de6c1ba90757d640c1c63c6339
          force: false
          localBranch: false
          tag: ""
        url: https://opendev.org/airship/treasuremap.git
    targetPath: /home/airship
permissions:
  DirectoryPermission: 488
  FilePermission: 416
>>>>>>>>Sample output of document pull for above configuration<<<<<<<<<
pkg/document/pull/pull.go:36: Reading current context manifest information from /home/airship/.airship/config
(currentContext:)
pkg/document/pull/pull.go:51: Downloading airshipctl repository airshipctl from https://opendev.org/airship/
airshipctl.git into /home/airship (url: & targetPath:)
pkg/document/repo/repo.go:141: Attempting to download the repository airshipctl
pkg/document/repo/repo.go:126: Attempting to clone the repository airshipctl from https://opendev.org/airship/
airshipctl.git
pkg/document/repo/repo.go:120: Attempting to open repository airshipctl
pkg/document/repo/repo.go:110: Attempting to checkout the repository airshipctl from commit hash #####
pkg/document/pull/pull.go:51: Downloading primary repository treasuremap from https://opendev.org/airship/
treasuremap.git into /home/airship  (repository name taken from url path last content)
pkg/document/repo/repo.go:141: Attempting to download the repository treasuremap
pkg/document/repo/repo.go:126: Attempting to clone the repository treasuremap from /home/airship/treasuremap
pkg/document/repo/repo.go:120: Attempting to open repository treasuremap
pkg/document/repo/repo.go:110: Attempting to checkout the repository treasuremap from commit hash #####
`
)

// NewPullCommand creates a new command for pulling airship document repositories
func NewPullCommand(cfgFactory config.Factory) *cobra.Command {
	var noCheckout bool
	documentPullCmd := &cobra.Command{
		Use: "pull",
		Long: fmt.Sprintf(long[1:], filepath.Join(
			config.HomeEnvVar, config.AirshipConfigDir, config.AirshipDefaultManifest)),
		Short:   "Airshipctl command to pull manifests from remote git repositories",
		Example: pullExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pull.Pull(cfgFactory, noCheckout)
		},
	}

	documentPullCmd.Flags().BoolVarP(&noCheckout, "no-checkout", "n", false,
		"no checkout is performed after the clone is complete.")

	return documentPullCmd
}

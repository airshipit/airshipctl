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

package cluster

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewClusterCommand returns cobra command object of the airshipctl cluster and adds it's subcommands.
func NewClusterCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	clusterRootCmd := &cobra.Command{
		Use: "cluster",
		// TODO: (kkalynovskyi) Add more description when more subcommands are added
		Short: "Control Kubernetes cluster",
		Long:  "Interactions with Kubernetes cluster, such as get status, deploy initial infrastructure",
	}

	clusterRootCmd.AddCommand(NewCmdInitInfra(rootSettings))

	return clusterRootCmd
}

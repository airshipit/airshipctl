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

	"opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	getKubeconfigLong = `
Retrieves kubeconfig of the cluster and prints it to stdout.

If you specify CLUSTER_NAME, kubeconfig will have a CurrentContext set to CLUSTER_NAME and
will have its context defined.

If you don't specify CLUSTER_NAME, kubeconfig will have multiple contexts for every cluster
in the airship site. Context names will correspond to cluster names. CurrentContext will be empty.
`
	getKubeconfigExample = `
Retrieve target-cluster kubeconfig
# airshipctl cluster get-kubeconfig target-cluster

Retrieve kubeconfig for the entire site; the kubeconfig will have context for every cluster
# airshipctl cluster get-kubeconfig
`
)

// NewGetKubeconfigCommand creates a command which retrieves cluster kubeconfig
func NewGetKubeconfigCommand(cfgFactory config.Factory) *cobra.Command {
	opts := &cluster.GetKubeconfigCommand{}
	cmd := &cobra.Command{
		Use:     "get-kubeconfig CLUSTER_NAME",
		Short:   "Airshipctl command to retrieve kubeconfig for a desired cluster",
		Long:    getKubeconfigLong[1:],
		Args:    GetKubeconfArgs(opts),
		Example: getKubeconfigExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.RunE(cfgFactory, cmd.OutOrStdout())
		},
	}
	return cmd
}

// GetKubeconfArgs extracts one or less arguments from command line, and saves it as name
func GetKubeconfArgs(opts *cluster.GetKubeconfigCommand) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			opts.ClusterName = args[0]
		}
		return cobra.MaximumNArgs(1)(cmd, args)
	}
}

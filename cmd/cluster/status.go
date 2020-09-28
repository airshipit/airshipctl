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
	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

// NewStatusCommand creates a command which reports the statuses of a cluster's deployed components.
func NewStatusCommand(cfgFactory config.Factory, factory client.Factory) *cobra.Command {
	o := cluster.NewStatusOptions(cfgFactory, factory)
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve statuses of deployed cluster components",
		RunE:  clusterStatusRunE(o),
	}
	return cmd
}

func clusterStatusRunE(o cluster.StatusOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return cluster.StatusRunner(o, cmd.OutOrStdout())
	}
}

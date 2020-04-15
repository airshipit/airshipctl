/*l
Copyright 2014 The Kubernetes Authors.

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
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	getClusterLong = `
Display a specific cluster or all defined clusters if no name is provided.

Note that if a specific cluster's name is provided, the --cluster-type flag
must also be provided.
Valid values for the --cluster-type flag are [ephemeral|target].
`

	getClusterExample = `
# List all clusters
airshipctl config get-cluster

# Display a specific cluster
airshipctl config get-cluster --cluster-type=ephemeral exampleCluster
`
)

// NewGetClusterCommand creates a command for viewing the cluster information
// defined in the airshipctl config file.
func NewGetClusterCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.ClusterOptions{}
	cmd := &cobra.Command{
		Use:     "get-cluster [NAME]",
		Short:   "Get cluster information from the airshipctl config",
		Long:    getClusterLong[1:],
		Example: getClusterExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			airconfig := rootSettings.Config
			if len(args) == 1 {
				o.Name = args[0]

				err := validate(o)
				if err != nil {
					return err
				}

				cluster, err := airconfig.GetCluster(o.Name, o.ClusterType)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), cluster.PrettyString())
				return nil
			}

			clusters := airconfig.GetClusters()
			if len(clusters) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No clusters found in the configuration.")
			}
			for _, cluster := range clusters {
				fmt.Fprintln(cmd.OutOrStdout(), cluster.PrettyString())
			}
			return nil
		},
	}

	addGetClusterFlags(o, cmd)
	return cmd
}

func addGetClusterFlags(o *config.ClusterOptions, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVar(
		&o.ClusterType,
		"cluster-type",
		"",
		"type of the desired cluster")
}

func validate(o *config.ClusterOptions) error {
	// Only an error if asking for a specific cluster
	if len(o.Name) == 0 {
		return nil
	}
	return config.ValidClusterType(o.ClusterType)
}

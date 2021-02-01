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

	"opendev.org/airship/airshipctl/pkg/clusterctl/client"
	clusterctlcmd "opendev.org/airship/airshipctl/pkg/clusterctl/cmd"
	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	getKubeconfigLong = `
Retrieve cluster kubeconfig and print it to stdout
`
	getKubeconfigExample = `
# Retrieve target-cluster kubeconfig
airshipctl cluster get-kubeconfig target-cluster --kubeconfig /tmp/kubeconfig
`
)

// NewGetKubeconfigCommand creates a command which retrieves cluster kubeconfig
func NewGetKubeconfigCommand(cfgFactory config.Factory) *cobra.Command {
	o := &client.GetKubeconfigOptions{}
	cmd := &cobra.Command{
		Use:     "get-kubeconfig [cluster_name]",
		Short:   "Retrieve kubeconfig for a desired cluster",
		Long:    getKubeconfigLong[1:],
		Example: getKubeconfigExample[1:],
		Args:    cobra.ExactArgs(1),
		RunE:    getKubeconfigRunE(cfgFactory, o),
	}

	initFlags(o, cmd)

	return cmd
}

func initFlags(o *client.GetKubeconfigOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.ParentKubeconfigPath,
		"kubeconfig",
		"",
		"path to kubeconfig associated with parental cluster")

	flags.StringVarP(
		&o.ManagedClusterNamespace,
		"namespace",
		"n",
		"default",
		"namespace where cluster is located, if not specified default one will be used")

	flags.StringVar(
		&o.ParentKubeconfigContext,
		"context",
		"",
		"specify context within the kubeconfig file")
}

// getKubeconfigRunE returns a function to cobra command to be executed in runtime
func getKubeconfigRunE(cfgFactory config.Factory, o *client.GetKubeconfigOptions) func(
	cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		o.ManagedClusterName = args[0]
		return clusterctlcmd.GetKubeconfig(cfgFactory, o, cmd.OutOrStdout())
	}
}

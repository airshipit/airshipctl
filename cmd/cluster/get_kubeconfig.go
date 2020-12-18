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

	"opendev.org/airship/airshipctl/pkg/errors"
)

const (
	getKubeconfigLong = `
Retrieve cluster kubeconfig and save it to file or stdout.
`
	getKubeconfigExample = `
# Retrieve target-cluster kubeconfig and print it to stdout
airshipctl cluster get-kubeconfig target-cluster
`
)

// NewGetKubeconfigCommand creates a command which retrieves cluster kubeconfig
func NewGetKubeconfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-kubeconfig [cluster_name]",
		Short:   "Retrieve kubeconfig for a desired cluster",
		Long:    getKubeconfigLong[1:],
		Example: getKubeconfigExample[1:],
		Args:    cobra.ExactArgs(1),
		RunE:    getKubeconfigRunE(),
	}

	return cmd
}

// getKubeconfigRunE returns a function to cobra command to be executed in runtime
func getKubeconfigRunE() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return errors.ErrNotImplemented{What: "cluster get-kubeconfig is not implemented yet"}
	}
}

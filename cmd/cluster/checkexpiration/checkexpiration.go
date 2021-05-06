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

package checkexpiration

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/cluster/checkexpiration"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	checkLong = `
Displays a list of certificate along with expirations from both the management and workload clusters, or in a
self-managed cluster. Checks for TLS Secrets, kubeconf secrets (which gets created while creating the
workload cluster) and also the node certificates present inside /etc/kubernetes/pki directory for each node.
`

	checkExample = `
To display all the expiring entities in the cluster
# airshipctl cluster check-certificate-expiration --kubeconfig testconfig

To display the entities whose expiration is within threshold of 30 days
# airshipctl cluster check-certificate-expiration -t 30 --kubeconfig testconfig

To output the contents to json (default operation)
# airshipctl cluster check-certificate-expiration -o json --kubeconfig testconfig
or
# airshipctl cluster check-certificate-expiration --kubeconfig testconfig

To output the contents to yaml
# airshipctl cluster check-certificate-expiration -o yaml --kubeconfig testconfig

To output the contents whose expiration is within 30 days to yaml
# airshipctl cluster check-certificate-expiration -t 30 -o yaml --kubeconfig testconfig
`

	kubeconfigFlag = "kubeconfig"
)

// NewCheckCommand creates a new command for generating secret information
func NewCheckCommand(cfgFactory config.Factory) *cobra.Command {
	c := &checkexpiration.CheckCommand{
		Options:       checkexpiration.CheckFlags{},
		CfgFactory:    cfgFactory,
		ClientFactory: client.DefaultClient,
	}

	checkCmd := &cobra.Command{
		Use: "check-certificate-expiration",
		Short: "Airshipctl command to check expiring TLS certificates, " +
			"secrets and kubeconfigs in the kubernetes cluster",
		Long:    checkLong[1:],
		Example: checkExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.RunE(cmd.OutOrStdout())
		},
	}

	checkCmd.Flags().StringVarP(&c.Options.FormatType, "output", "o", "json", "convert output to yaml or json")
	checkCmd.Flags().StringVar(&c.Options.KubeContext, "kubecontext", "", "kubeconfig context to be used")
	checkCmd.Flags().StringVar(&c.Options.Kubeconfig, kubeconfigFlag, "",
		"path to kubeconfig associated with cluster being managed")
	checkCmd.Flags().IntVarP(&c.Options.Threshold, "threshold", "t", -1,
		"the max expiration threshold in days before a certificate is expiring. Displays all the certificates by default")

	err := checkCmd.MarkFlagRequired(kubeconfigFlag)
	if err != nil {
		log.Fatalf("marking kubeconfig flag required failed: %v", err)
	}
	return checkCmd
}

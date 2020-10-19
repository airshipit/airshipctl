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

package resetsatoken

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/cluster/resetsatoken"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	resetLong = `
Use to reset/rotate the Service Account(SA) tokens and additionally restart the
corresponding pods to get the latest token data reflected in the pod spec

Secret-namespace is a mandatory field and secret-name is optional. If secret-
name is not given, all the SA tokens in that particular namespace is considered,
else only that particular input secret-name`

	resetExample = `
# To rotate a particular SA token
airshipctl cluster rotate-sa-token -n cert-manager -s cert-manager-token-vvn9p

# To rotate all the SA tokens in cert-manager namespace
airshipctl cluster rotate-sa-token -n cert-manager
`
)

// NewResetCommand creates a new command for generating secret information
func NewResetCommand(cfgFactory config.Factory) *cobra.Command {
	r := &resetsatoken.ResetCommand{
		Options:    resetsatoken.ResetFlags{},
		CfgFactory: cfgFactory,
	}

	resetCmd := &cobra.Command{
		Use:     "rotate-sa-token",
		Short:   "Rotate tokens of Service Accounts",
		Long:    resetLong[1:],
		Example: resetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return r.RunE()
		},
	}

	resetCmd.Flags().StringVarP(&r.Options.Namespace, "secret-namespace", "n", "",
		"namespace of the Service Account Token")
	resetCmd.Flags().StringVarP(&r.Options.SecretName, "secret-name", "s", "",
		"name of the secret containing Service Account Token")
	resetCmd.Flags().StringVar(&r.Options.Kubeconfig, "kubeconfig", "",
		"Path to kubeconfig associated with cluster being managed")

	err := resetCmd.MarkFlagRequired("secret-namespace")
	if err != nil {
		log.Fatal(err)
	}
	err = resetCmd.MarkFlagRequired("kubeconfig")
	if err != nil {
		log.Fatalf("marking kubeconfig flag required failed: %v", err)
	}
	return resetCmd
}

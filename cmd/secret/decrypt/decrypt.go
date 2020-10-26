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

package decrypt

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/secret"
)

const (
	decryptShort = `
Decrypt encrypted yaml files into plaintext files representing Kubernetes objects consisting of sensitive data.`

	decryptExample = `
# Decrypt all encrypted files in the manifests directory.
airshipctl secret decrypt

# Decrypt encrypted file from src and write the plain text to a different dst file
airshipctl secret decrypt \
	--src /tmp/manifests/target/secrets/encrypted-qualified-secret.yaml \
	--dst /tmp/manifests/target/secrets/qualified-secret.yaml
`
)

// NewDecryptCommand creates a new command for decrypting encrypted secrets in the manifests
func NewDecryptCommand(cfgFactory config.Factory) *cobra.Command {
	var srcPath, dstPath, kubeconfig string

	decryptCmd := &cobra.Command{
		Use:     "decrypt",
		Short:   decryptShort[1:],
		Example: decryptExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			airshipConfig, err := cfgFactory()
			if err != nil {
				return err
			}
			return secret.Decrypt(airshipConfig, kubeconfig, srcPath, dstPath)
		},
	}

	decryptCmd.Flags().StringVar(&srcPath, "src", "",
		`Path to the file or directory that has secrets in encrypted text that need to be decrypted. `+
			`Defaults to the manifest location in airship config`)
	decryptCmd.Flags().StringVar(&dstPath, "dst", "",
		"Path to the file or directory to store decrypted secrets. Defaults to src if empty.")
	decryptCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "",
		"Path to kubeconfig associated with cluster being managed")

	err := decryptCmd.MarkFlagRequired("dst")
	if err != nil {
		log.Fatalf("marking dst flag required failed: %v", err)
	}

	return decryptCmd
}

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

package encrypt

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	encryptShort = `
Encrypt plain text yaml files representing Kubernetes objects consisting of sensitive configuration.`

	encryptExample = `
# Encrypt all kubernetes objects in the manifests directory.
airshipctl secret encrypt

# Encrypt file from src and write to a different dst file
airshipctl secret encrypt \
	--src /tmp/manifests/target/secrets/qualified-secret.yaml \
	--dst /tmp/manifests/target/secrets/encrypted-qualified-secret.yaml
`
)

// NewEncryptCommand creates a new command for encrypting plain text secrets using sops
func NewEncryptCommand(_ config.Factory) *cobra.Command {
	var srcPath, dstPath string

	encryptCmd := &cobra.Command{
		Use:     "encrypt",
		Short:   encryptShort[1:],
		Example: encryptExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.ErrNotImplemented{What: "secret encryption/decryption"}
		},
	}
	encryptCmd.Flags().StringVar(&srcPath, "src", "",
		`Path to the file or directory that has secrets in plaintext that need to be encrypted. `+
			`Defaults to the manifest location in airship config`)
	encryptCmd.Flags().StringVar(&dstPath, "dst", "",
		"Path to the file or directory that has encrypted secrets for decryption. Defaults to src if empty.")

	err := encryptCmd.MarkFlagRequired("dst")
	if err != nil {
		log.Fatalf("marking dst flag required failed: %v", err)
	}

	return encryptCmd
}

/*
Copyright 2016 The Kubernetes Authors.

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
)

const (
	setEncryptionConfigLong = `
Create or modify an encryption config in the airshipctl config file.

Encryption configs are local files or kubernetes secrets that are used to encrypt and decrypt kubernetes objects
`

	setEncryptionConfigExample = `
# Create an encryption config with local gpg key source
airshipctl config set-encryption-config exampleConfig \
  --encryption-key path-to-encryption-key \
  --decryption-key path-to-encryption-key

# Create an encryption config with kube api server secret as the store to store encryption keys
airshipctl config set-encryption-config exampleConfig \
  --secret-name secretName \
  --secret-namespace secretNamespace
`
)

// NewSetEncryptionConfigCommand creates a command for creating and modifying encryption
// configs in the airshipctl config file.
func NewSetEncryptionConfigCommand(cfgFactory config.Factory) *cobra.Command {
	o := &config.EncryptionConfigOptions{}
	cmd := &cobra.Command{
		Use:     "set-encryption-config NAME",
		Short:   "Manage encryption configs in airship config",
		Long:    setEncryptionConfigLong[1:],
		Example: setEncryptionConfigExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgFactory()
			if err != nil {
				return err
			}
			o.Name = args[0]
			modified, err := config.RunSetEncryptionConfig(o, cfg, true)
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Encryption Config %q modified.\n", o.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Encryption Config %q created.\n", o.Name)
			}
			return nil
		},
	}

	addSetEncryptionConfigFlags(o, cmd)
	return cmd
}

func addSetEncryptionConfigFlags(o *config.EncryptionConfigOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.EncryptionKeyPath,
		"encryption-key-path",
		"",
		"the path to the encryption key file")

	flags.StringVar(
		&o.DecryptionKeyPath,
		"decryption-key-path",
		"",
		"the path to the decryption key file")

	flags.StringVar(
		&o.KeySecretName,
		"secret-name",
		"",
		"name of the secret consisting of the encryption and decryption keys")

	flags.StringVar(
		&o.KeySecretNamespace,
		"secret-namespace",
		"",
		"namespace of the secret consisting of the encryption and decryption keys")
}

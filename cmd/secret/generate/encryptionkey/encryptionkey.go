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

package encryptionkey

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/secret/generate"
)

const (
	cmdLong = `
Generates a secure encryption key or passphrase.

If regex arguments are passed the encryption key created would match the regular expression passed.
`

	cmdExample = `
# Generates a secure encryption key or passphrase.
airshipctl secret generate encryptionkey

# Generates a secure encryption key or passphrase matching the regular expression
airshipctl secret generate encryptionkey \
  --regex Xy[a-c][0-9]!a*
`
)

// NewGenerateEncryptionKeyCommand creates a new command for generating secret information
func NewGenerateEncryptionKeyCommand() *cobra.Command {
	var regex string
	var limit int

	encryptionKeyCmd := &cobra.Command{
		Use:     "encryptionkey",
		Short:   "Generates a secure encryption key or passphrase",
		Long:    cmdLong[1:],
		Example: cmdExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("limit") && !cmd.Flags().Changed("regex") {
				return fmt.Errorf("required Regex flag with limit option")
			}
			if cmd.Flags().Changed("regex") && cmd.Flags().Changed("limit") {
				return errors.ErrNotImplemented{What: "Regex support not implemented yet!"}
			}
			engine := generate.NewEncryptionKeyEngine(nil)
			encryptionKey := engine.GenerateEncryptionKey()
			fmt.Fprintln(cmd.OutOrStdout(), encryptionKey)
			return nil
		},
	}

	encryptionKeyCmd.Flags().StringVar(&regex, "regex", "",
		"Regular expression string")

	encryptionKeyCmd.Flags().IntVar(&limit, "limit", 5,
		"Limit number of characters for + or * regex")
	return encryptionKeyCmd
}

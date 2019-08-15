package generate

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/secret"
)

// NewGenerateMasterPassphraseCommand creates a new command for generating secret information
func NewGenerateMasterPassphraseCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	masterPassphraseCmd := &cobra.Command{
		Use: "masterpassphrase",
		// TODO(howell): Make this more expressive
		Short: "generates a secure master passphrase",
		Run: func(cmd *cobra.Command, args []string) {
			engine := secret.NewPassphraseEngine(nil)
			masterPassphrase := engine.GeneratePassphrase()
			fmt.Fprintln(cmd.OutOrStdout(), masterPassphrase)
		},
	}

	return masterPassphraseCmd
}

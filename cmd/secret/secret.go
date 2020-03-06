package secret

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/cmd/secret/generate"
)

// NewSecretCommand creates a new command for managing airshipctl secrets
func NewSecretCommand() *cobra.Command {
	secretRootCmd := &cobra.Command{
		Use: "secret",
		// TODO(howell): Make this more expressive
		Short: "manages secrets",
	}

	secretRootCmd.AddCommand(generate.NewGenerateCommand())

	return secretRootCmd
}

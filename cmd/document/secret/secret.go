package secret

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/cmd/document/secret/generate"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewSecretCommand creates a new command for managing airshipctl secrets
func NewSecretCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	secretRootCmd := &cobra.Command{
		Use: "secret",
		// TODO(howell): Make this more expressive
		Short: "manages secrets",
	}

	secretRootCmd.AddCommand(generate.NewGenerateCommand(rootSettings))

	return secretRootCmd
}

package generate

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewGenerateCommand creates a new command for generating secret information
func NewGenerateCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	generateRootCmd := &cobra.Command{
		Use: "generate",
		// TODO(howell): Make this more expressive
		Short: "generates various secrets",
	}

	generateRootCmd.AddCommand(NewGenerateMasterPassphraseCommand(rootSettings))

	return generateRootCmd
}

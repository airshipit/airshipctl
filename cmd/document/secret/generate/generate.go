package generate

import "github.com/spf13/cobra"

// NewGenerateCommand creates a new command for generating secret information
func NewGenerateCommand() *cobra.Command {
	generateRootCmd := &cobra.Command{
		Use: "generate",
		// TODO(howell): Make this more expressive
		Short: "generates various secrets",
	}

	generateRootCmd.AddCommand(NewGenerateMasterPassphraseCommand())

	return generateRootCmd
}

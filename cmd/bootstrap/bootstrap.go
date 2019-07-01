package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/pkg/environment"
)

// NewBootstrapCommand creates a new command for bootstrapping airshipctl
func NewBootstrapCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	bootstrapRootCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "bootstraps airshipctl",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Under construction\n")
		},
	}

	return bootstrapRootCmd
}

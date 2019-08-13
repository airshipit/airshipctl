package bootstrap

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewBootstrapCommand creates a new command for bootstrapping airshipctl
func NewBootstrapCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	bootstrapRootCmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap ephemeral Kubernetes cluster",
	}

	ISOGenCmd := NewISOGenCommand(bootstrapRootCmd, rootSettings)
	bootstrapRootCmd.AddCommand(ISOGenCmd)

	return bootstrapRootCmd
}

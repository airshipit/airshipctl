package bootstrap

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// New Bootstrap remote direct subcommand
func NewRemoteDirectCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	remoteDirect := &cobra.Command{
		Use:   "remotedirect",
		Short: "Bootstrap ephemeral node",
		RunE: func(cmd *cobra.Command, args []string) error {
			return remote.DoRemoteDirect(rootSettings)
		},
	}

	return remoteDirect
}

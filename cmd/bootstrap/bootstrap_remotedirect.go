package bootstrap

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// NewRemoteDirectCommand provides a command with the capability to perform remote direct operations.
func NewRemoteDirectCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	remoteDirect := &cobra.Command{
		Use:   "remotedirect",
		Short: "Bootstrap ephemeral node",
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := remote.NewAdapter(rootSettings)
			if err != nil {
				return err
			}

			return a.DoRemoteDirect()
		},
	}

	return remoteDirect
}

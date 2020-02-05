package bootstrap

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	alog "opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote"
)

// RemoteDirect settings for remotedirect command
type RemoteDirectSettings struct {
	*environment.AirshipCTLSettings

	RemoteConfig remote.RemoteDirectConfig
}

// InitFlags adds flags and their default settings for Remote Direct command
func (cmdSetting *RemoteDirectSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.Flags()

	// TODO: Remove CLI flags after reading configuration from config documents
	// ========================================================================
	flags.StringVar(&cmdSetting.RemoteConfig.RemoteURL,
		"remote-url",
		"http://localhost:8000",
		"[Temporary. Will be removed] Remote Redfish/Smash URL")

	flags.StringVar(&cmdSetting.RemoteConfig.EphemeralNodeId,
		"eph-node-id",
		"",
		"[Temporary. Will be removed] Ephemeral Node ID")

	flags.StringVar(&cmdSetting.RemoteConfig.IsoPath,
		"iso-path",
		"",
		"[Temporary. Will be removed] Remote ISO path for ephemeral node")

	flags.StringVar(&cmdSetting.RemoteConfig.RemoteType,
		"remote-type",
		"redfish",
		"Remote type e.g. redfish, smash etc.")

	err := cmd.MarkFlagRequired("eph-node-id")
	if err != nil {
		alog.Fatal(err)
	}

	err = cmd.MarkFlagRequired("iso-path")
	if err != nil {
		alog.Fatal(err)
	}
}

// New Bootstrap remote direct subcommand
func NewRemoteDirectCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	settings := &RemoteDirectSettings{AirshipCTLSettings: rootSettings}
	remoteDirect := &cobra.Command{
		Use:   "remotedirect",
		Short: "Bootstrap ephemeral node",
		RunE: func(cmd *cobra.Command, args []string) error {
			/* TODO: Get config from settings.GetCurrentContext() and remove cli arguments */

			/* Trigger remotedirect based on remote type */
			return remote.DoRemoteDirect(settings.RemoteConfig)
		},
	}

	settings.InitFlags(remoteDirect)

	return remoteDirect
}

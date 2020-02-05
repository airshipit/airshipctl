package remote

import (
	"context"

	alog "opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	AirshipRemoteTypeRedfish string = "redfish"
	AirshipRemoteTypeSmash   string = "smash"
)

// This structure defines the common remote direct config
// for all remote types.
type RemoteDirectConfig struct {
	// remote type
	RemoteType string

	// remote URL
	RemoteURL string

	// ephemeral Host ID
	EphemeralNodeId string

	// ISO URL
	IsoPath string

	// TODO: Ephemeral Node IP

	// TODO: kubeconfig (in object form or raw yaml?) for ephemeral node validation.

	// TODO: More fields can be added on need basis
}

// Interface to be implemented by remoteDirect implementation
type RemoteDirectClient interface {
	DoRemoteDirect() error
}

// Get remotedirect client based on config
func getRemoteDirectClient(remoteConfig RemoteDirectConfig) (RemoteDirectClient, error) {
	var client RemoteDirectClient
	var err error

	switch remoteConfig.RemoteType {
	case AirshipRemoteTypeRedfish:
		alog.Debug("Remote type redfish")

		client, err = redfish.NewRedfishRemoteDirectClient(
			context.Background(),
			remoteConfig.RemoteURL,
			remoteConfig.EphemeralNodeId,
			remoteConfig.IsoPath)
		if err != nil {
			alog.Debugf("redfish remotedirect client creation failed")
			return nil, err
		}

	default:
		return nil, NewRemoteDirectErrorf("invalid remote type")
	}

	return client, nil
}

// Top level function to execute remote direct based on remote type
func DoRemoteDirect(remoteConfig RemoteDirectConfig) error {
	client, err := getRemoteDirectClient(remoteConfig)
	if err != nil {
		return err
	}

	err = client.DoRemoteDirect()
	if err != nil {
		alog.Debugf("remote direct failed: %s", err)
		return err
	}

	alog.Print("Remote direct successfully completed")

	return nil
}

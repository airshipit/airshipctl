package remote

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	alog "opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	AirshipRemoteTypeRedfish string = "redfish"
	AirshipHostKind          string = "BareMetalHost"
)

// Interface to be implemented by remoteDirect implementation
type Client interface {
	DoRemoteDirect() error
}

// Get remotedirect client based on config
func getRemoteDirectClient(remoteConfig *config.RemoteDirect, remoteURL string) (Client, error) {
	var client Client
	switch remoteConfig.RemoteType {
	case AirshipRemoteTypeRedfish:
		alog.Debug("Remote type redfish")

		rfURL, err := url.Parse(remoteURL)
		if err != nil {
			return nil, err
		}

		baseURL := fmt.Sprintf("%s://%s", rfURL.Scheme, rfURL.Host)
		schemeSplit := strings.Split(rfURL.Scheme, redfish.RedfishURLSchemeSeparator)
		if len(schemeSplit) > 1 {
			baseURL = fmt.Sprintf("%s://%s", schemeSplit[len(schemeSplit)-1], rfURL.Host)
		}

		urlPath := strings.Split(rfURL.Path, "/")
		nodeID := urlPath[len(urlPath)-1]

		client, err = redfish.NewRedfishRemoteDirectClient(
			context.Background(),
			baseURL,
			nodeID,
			remoteConfig.IsoURL)
		if err != nil {
			alog.Debugf("redfish remotedirect client creation failed")
			return nil, err
		}

	default:
		return nil, NewRemoteDirectErrorf("invalid remote type")
	}

	return client, nil
}

func getRemoteDirectConfig(settings *environment.AirshipCTLSettings) (*config.RemoteDirect, string, error) {
	cfg := settings.Config()
	manifest, err := cfg.CurrentContextManifest()
	if err != nil {
		return nil, "", err
	}
	bootstrapSettings, err := cfg.CurrentContextBootstrapInfo()
	if err != nil {
		return nil, "", err
	}

	remoteConfig := bootstrapSettings.RemoteDirect
	if remoteConfig == nil {
		return nil, "", config.ErrMissingConfig{What: "RemoteDirect options not defined in bootstrap config"}
	}

	// TODO (dukov) replace with the appropriate function once it's available
	// in document module
	docBundle, err := document.NewBundle(document.NewDocumentFs(), manifest.TargetPath, "")
	if err != nil {
		return nil, "", err
	}

	ls := document.EphemeralClusterSelector
	selector := document.NewSelector().
		ByGvk("", "", AirshipHostKind).
		ByLabel(ls)
	docs, err := docBundle.Select(selector)
	if err != nil {
		return nil, "", err
	}
	if len(docs) == 0 {
		return nil, "", document.ErrDocNotFound{
			Selector: ls,
			Kind:     AirshipHostKind,
		}
	}

	// NOTE If filter returned more than one document chose first
	remoteURL, err := docs[0].GetString("spec.bmc.address")
	if err != nil {
		return nil, "", err
	}

	return remoteConfig, remoteURL, nil
}

// Top level function to execute remote direct based on remote type
func DoRemoteDirect(settings *environment.AirshipCTLSettings) error {
	remoteConfig, remoteURL, err := getRemoteDirectConfig(settings)
	if err != nil {
		return err
	}

	client, err := getRemoteDirectClient(remoteConfig, remoteURL)
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

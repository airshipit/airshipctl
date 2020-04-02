package remote

import (
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
)

// Interface to be implemented by remoteDirect implementation
type RDClient interface {
	DoRemoteDirect() error
}

// Get remotedirect client based on config
func getRemoteDirectClient(
	remoteConfig *config.RemoteDirect,
	remoteURL string,
	username string,
	password string) (RDClient, error) {
	var client RDClient
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
			baseURL,
			nodeID,
			username,
			password,
			remoteConfig.IsoURL,
			remoteConfig.Insecure,
			remoteConfig.UseProxy,
		)
		if err != nil {
			alog.Debugf("redfish remotedirect client creation failed")
			return nil, err
		}

	default:
		return nil, NewRemoteDirectErrorf("invalid remote type")
	}

	return client, nil
}

func getRemoteDirectConfig(settings *environment.AirshipCTLSettings) (
	remoteConfig *config.RemoteDirect,
	remoteURL string,
	username string,
	password string,
	err error) {
	cfg := settings.Config()
	bootstrapSettings, err := cfg.CurrentContextBootstrapInfo()
	if err != nil {
		return nil, "", "", "", err
	}

	remoteConfig = bootstrapSettings.RemoteDirect
	if remoteConfig == nil {
		return nil, "", "", "", config.ErrMissingConfig{What: "RemoteDirect options not defined in bootstrap config"}
	}

	bundlePath, err := cfg.CurrentContextEntryPoint(config.Ephemeral, "")
	if err != nil {
		return nil, "", "", "", err
	}

	docBundle, err := document.NewBundleByPath(bundlePath)
	if err != nil {
		return nil, "", "", "", err
	}

	selector := document.NewEphemeralBMHSelector()
	doc, err := docBundle.SelectOne(selector)
	if err != nil {
		return nil, "", "", "", err
	}

	remoteURL, err = document.GetBMHBMCAddress(doc)
	if err != nil {
		return nil, "", "", "", err
	}

	username, password, err = document.GetBMHBMCCredentials(doc, docBundle)
	if err != nil {
		return nil, "", "", "", err
	}

	return remoteConfig, remoteURL, username, password, nil
}

// Top level function to execute remote direct based on remote type
func DoRemoteDirect(settings *environment.AirshipCTLSettings) error {
	remoteConfig, remoteURL, username, password, err := getRemoteDirectConfig(settings)
	if err != nil {
		return err
	}

	client, err := getRemoteDirectClient(remoteConfig, remoteURL, username, password)
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

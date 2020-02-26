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

	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

const (
	AirshipRemoteTypeRedfish string = "redfish"
	AirshipRemoteTypeSmash   string = "smash"
	AirshipHostKind          string = "BareMetalHost"
)

// Interface to be implemented by remoteDirect implementation
type RemoteDirectClient interface {
	DoRemoteDirect() error
}

// Get remotedirect client based on config
func getRemoteDirectClient(remoteConfig *config.RemoteDirect, remoteURL string) (RemoteDirectClient, error) {
	var client RemoteDirectClient
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
	docBundle, err := document.NewBundle(fs.MakeRealFS(), manifest.TargetPath, "")
	if err != nil {
		return nil, "", err
	}

	label := document.EphemeralClusterSelector
	filter := types.Selector{
		Gvk:           gvk.FromKind(AirshipHostKind),
		LabelSelector: label,
	}
	docs, err := docBundle.Select(filter)
	if err != nil {
		return nil, "", err
	}
	if len(docs) == 0 {
		return nil, "", document.ErrDocNotFound{
			Selector: label,
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

package redfish

import (
	"context"
	"fmt"
	"net/url"

	redfishApi "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	alog "opendev.org/airship/airshipctl/pkg/log"
)

type RedfishRemoteDirect struct {

	// Context
	Context context.Context

	// remote URL
	RemoteURL url.URL

	// ephemeral Host ID
	EphemeralNodeID string

	// ISO URL
	IsoPath string

	// Redfish Client implementation
	RedfishAPI redfishApi.RedfishAPI
}

// Top level function to handle Redfish remote direct
func (cfg RedfishRemoteDirect) DoRemoteDirect() error {
	alog.Debugf("Using Redfish Endpoint: '%s'", cfg.RemoteURL.String())

	/* TODO: Add Authentication when redfish library supports it. */

	/* Get system details */
	systemID := cfg.EphemeralNodeID
	system, _, err := cfg.RedfishAPI.GetSystem(cfg.Context, systemID)
	if err != nil {
		return NewRedfishClientErrorf("Get System[%s] failed with err: %s", systemID, err.Error())
	}
	alog.Debugf("Ephemeral Node System ID: '%s'", systemID)

	/* get manager for system */
	managerID := GetResourceIDFromURL(system.Links.ManagedBy[0].OdataId)
	alog.Debugf("Ephemeral node managerID: '%s'", managerID)

	/* Get manager's Cd or DVD virtual media ID */
	vMediaID, vMediaType, err := GetVirtualMediaID(cfg.Context, cfg.RedfishAPI, managerID)
	if err != nil {
		return err
	}
	alog.Debugf("Ephemeral Node Virtual Media Id: '%s'", vMediaID)

	/* Load ISO in manager's virtual media */
	err = SetVirtualMedia(cfg.Context, cfg.RedfishAPI, managerID, vMediaID, cfg.IsoPath)
	if err != nil {
		return err
	}
	alog.Debugf("Successfully loaded virtual media: '%s'", cfg.IsoPath)

	/* Set system's bootsource to selected media */
	err = SetSystemBootSourceForMediaType(cfg.Context, cfg.RedfishAPI, systemID, vMediaType)
	if err != nil {
		return err
	}

	/* Reboot system */
	err = RebootSystem(cfg.Context, cfg.RedfishAPI, systemID)
	if err != nil {
		return err
	}
	alog.Debug("Restarted ephemeral host")

	return nil
}

// Creates a new Redfish remote direct client.
func NewRedfishRemoteDirectClient(ctx context.Context,
	remoteURL string,
	ephNodeID string,
	isoPath string,
) (RedfishRemoteDirect, error) {
	if remoteURL == "" {
		return RedfishRemoteDirect{},
			ErrRedfishMissingConfig{
				What: "redfish remote url empty",
			}
	}

	if ephNodeID == "" {
		return RedfishRemoteDirect{},
			ErrRedfishMissingConfig{
				What: "redfish ephemeral node id empty",
			}
	}

	if isoPath == "" {
		return RedfishRemoteDirect{},
			ErrRedfishMissingConfig{
				What: "redfish ephemeral node iso Path empty",
			}
	}

	cfg := &redfishClient.Configuration{
		BasePath:      remoteURL,
		DefaultHeader: make(map[string]string),
		UserAgent:     "airshipctl/client",
	}

	var api redfishApi.RedfishAPI = redfishClient.NewAPIClient(cfg).DefaultApi

	parsedURL, err := url.Parse(remoteURL)
	if err != nil {
		return RedfishRemoteDirect{},
			ErrRedfishMissingConfig{
				What: fmt.Sprintf("invalid url format: %v", err),
			}
	}

	client := RedfishRemoteDirect{
		Context:         ctx,
		RemoteURL:       *parsedURL,
		EphemeralNodeID: ephNodeID,
		IsoPath:         isoPath,
		RedfishAPI:      api,
	}

	return client, nil
}

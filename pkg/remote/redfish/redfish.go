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
	EphemeralNodeId string

	// ISO URL
	IsoPath string

	// Redfish Client implementation
	Api redfishApi.RedfishAPI
}

// Top level function to handle Redfish remote direct
func (cfg RedfishRemoteDirect) DoRemoteDirect() error {
	alog.Debugf("Using Redfish Endpoint: '%s'", cfg.RemoteURL.String())

	/* TODO: Add Authentication when redfish library supports it. */

	/* Get system details */
	systemId := cfg.EphemeralNodeId
	system, _, err := cfg.Api.GetSystem(cfg.Context, systemId)
	if err != nil {
		return NewRedfishClientErrorf("Get System[%s] failed with err: %s", systemId, err.Error())
	}
	alog.Debugf("Ephemeral Node System ID: '%s'", systemId)

	/* get manager for system */
	managerId := GetResourceIDFromURL(system.Links.ManagedBy[0].OdataId)
	alog.Debugf("Ephemeral node managerId: '%s'", managerId)

	/* Get manager's Cd or DVD virtual media ID */
	vMediaId, vMediaType, err := GetVirtualMediaId(cfg.Context, cfg.Api, managerId)
	if err != nil {
		return err
	}
	alog.Debugf("Ephemeral Node Virtual Media Id: '%s'", vMediaId)

	/* Load ISO in manager's virtual media */
	err = SetVirtualMedia(cfg.Context, cfg.Api, managerId, vMediaId, cfg.IsoPath)
	if err != nil {
		return err
	}
	alog.Debugf("Successfully loaded virtual media: '%s'", cfg.IsoPath)

	/* Set system's bootsource to selected media */
	err = SetSystemBootSourceForMediaType(cfg.Context, cfg.Api, systemId, vMediaType)
	if err != nil {
		return err
	}

	/* Reboot system */
	err = RebootSystem(cfg.Context, cfg.Api, systemId)
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

	parsedUrl, err := url.Parse(remoteURL)
	if err != nil {
		return RedfishRemoteDirect{},
			ErrRedfishMissingConfig{
				What: fmt.Sprintf("invalid url format: %v", err),
			}
	}

	client := RedfishRemoteDirect{
		Context:         ctx,
		RemoteURL:       *parsedUrl,
		EphemeralNodeId: ephNodeID,
		IsoPath:         isoPath,
		Api:             api,
	}

	return client, nil
}

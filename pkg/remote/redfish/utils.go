package redfish

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	redfishApi "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	SystemRebootDelay         = 2 * time.Second
	SystemActionRetries       = 30
	RedfishURLSchemeSeparator = "+"
)

// Redfish Id ref is a URI which contains resource Id
// as the last part. This function extracts resource
// ID from ID ref
func GetResourceIDFromURL(refURL string) string {
	u, err := url.Parse(refURL)
	if err != nil {
		log.Fatal(err)
	}

	trimmedURL := strings.TrimSuffix(u.Path, "/")
	elems := strings.Split(trimmedURL, "/")

	id := elems[len(elems)-1]
	return id
}

// Checks whether an ID exists in Redfish IDref collection
func IsIDInList(idRefList []redfishClient.IdRef, id string) bool {
	for _, r := range idRefList {
		rID := GetResourceIDFromURL(r.OdataId)
		if rID == id {
			return true
		}
	}
	return false
}

// GetVirtualMediaID retrieves the ID of a Redfish virtual media resource if it supports type "CD" or "DVD".
func GetVirtualMediaID(ctx context.Context, api redfishApi.RedfishAPI, managerID string) (string, string, error) {
	mediaCollection, httpResp, err := api.ListManagerVirtualMedia(ctx, managerID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return "", "", err
	}

	for _, mediaURI := range mediaCollection.Members {
		// Retrieve the virtual media ID from the request URI
		mediaID := GetResourceIDFromURL(mediaURI.OdataId)

		vMedia, httpResp, err := api.GetManagerVirtualMedia(ctx, managerID, mediaID)
		if err = ScreenRedfishError(httpResp, err); err != nil {
			return "", "", err
		}

		for _, mediaType := range vMedia.MediaTypes {
			if mediaType == "CD" || mediaType == "DVD" {
				return mediaID, mediaType, nil
			}
		}
	}

	return "", "", ErrRedfishClient{Message: "Unable to find virtual media with type CD or DVD"}
}

// This function walks through the bootsources of a system and sets the bootsource
// which is compatible with the given media type.
func SetSystemBootSourceForMediaType(ctx context.Context,
	api redfishApi.RedfishAPI,
	systemID string,
	mediaType string) error {
	/* Check available boot sources for system */
	system, _, err := api.GetSystem(ctx, systemID)
	if err != nil {
		return ErrRedfishClient{Message: fmt.Sprintf("Get System[%s] failed with err: %v", systemID, err)}
	}

	allowableValues := system.Boot.BootSourceOverrideTargetRedfishAllowableValues
	for _, bootSource := range allowableValues {
		if strings.EqualFold(string(bootSource), mediaType) {
			/* set boot source */
			systemReq := redfishClient.ComputerSystem{}
			systemReq.Boot.BootSourceOverrideTarget = bootSource
			_, httpResp, err := api.SetSystem(ctx, systemID, systemReq)
			return ScreenRedfishError(httpResp, err)
		}
	}

	return ErrRedfishClient{Message: fmt.Sprintf("failed to set system[%s] boot source", systemID)}
}

// Reboots a system by force shutoff and turning on.
func RebootSystem(ctx context.Context, api redfishApi.RedfishAPI, systemID string) error {
	waitForPowerState := func(desiredState redfishClient.PowerState) error {
		// Check if number of retries is defined in context
		totalRetries, ok := ctx.Value("numRetries").(int)
		if !ok {
			totalRetries = SystemActionRetries
		}

		for retry := 0; retry <= totalRetries; retry++ {
			system, httpResp, err := api.GetSystem(ctx, systemID)
			if err = ScreenRedfishError(httpResp, err); err != nil {
				return err
			}
			if system.PowerState == desiredState {
				return nil
			}
			time.Sleep(SystemRebootDelay)
		}
		return ErrOperationRetriesExceeded{}
	}

	resetReq := redfishClient.ResetRequestBody{}

	// Send PowerOff request
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	_, httpResp, err := api.ResetSystem(ctx, systemID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}
	// Check that node is powered off
	if err = waitForPowerState(redfishClient.POWERSTATE_OFF); err != nil {
		return err
	}

	// Send PowerOn request
	resetReq.ResetType = redfishClient.RESETTYPE_ON
	_, httpResp, err = api.ResetSystem(ctx, systemID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}
	// Check that node is powered on and return
	return waitForPowerState(redfishClient.POWERSTATE_ON)
}

// Insert the remote virtual media to the given virtual media id.
// This assumes that isoPath is accessible to the redfish server and
// virtualMedia device is either of type CD or DVD.
func SetVirtualMedia(ctx context.Context,
	api redfishApi.RedfishAPI,
	managerID string,
	vMediaID string,
	isoPath string) error {
	vMediaReq := redfishClient.InsertMediaRequestBody{}
	vMediaReq.Image = isoPath
	vMediaReq.Inserted = true
	_, httpResp, err := api.InsertVirtualMedia(ctx, managerID, vMediaID, vMediaReq)
	return ScreenRedfishError(httpResp, err)
}

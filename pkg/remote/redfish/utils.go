package redfish

import (
	"context"
	"net/url"
	"strings"
	"time"

	redfishApi "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	SystemRebootDelay         = 2 * time.Second
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

// This function walks through the list of manager's virtual media resources
// and gets the ID of virtualmedia which has type "CD" or "DVD"
func GetVirtualMediaID(ctx context.Context,
	api redfishApi.RedfishAPI,
	managerID string,
) (string, string, error) {
	// TODO: Sushy emulator has a bug which sends 'virtualMedia.inserted' field as
	//       string instead of a boolean which causes the redfish client to fail
	//       while unmarshalling this field.
	//       Insert logic here after the bug is fixed in sushy-emulator.
	return "Cd", "CD", nil
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
		return NewRedfishClientErrorf("Get System[%s] failed with err: %s", systemID, err.Error())
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

	return NewRedfishClientErrorf("failed to set system[%s] boot source", systemID)
}

// Reboots a system by force shutoff and turning on.
func RebootSystem(ctx context.Context, api redfishApi.RedfishAPI, systemID string) error {
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	_, httpResp, err := api.ResetSystem(ctx, systemID, resetReq)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return err
	}

	time.Sleep(SystemRebootDelay)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	_, httpResp, err = api.ResetSystem(ctx, systemID, resetReq)

	return ScreenRedfishError(httpResp, err)
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

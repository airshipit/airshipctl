// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	URLSchemeSeparator = "+"
)

// GetManagerID retrieves the manager ID for a redfish system.
func GetManagerID(ctx context.Context, api redfishAPI.RedfishAPI, systemID string) (string, error) {
	system, httpResp, err := api.GetSystem(ctx, systemID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		return "", err
	}

	return GetResourceIDFromURL(system.Links.ManagedBy[0].OdataId), nil
}

// GetResourceIDFromURL returns a parsed Redfish resource ID
// from the Redfish URL
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

// IsIDInList checks whether an ID exists in Redfish IDref collection
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
func GetVirtualMediaID(ctx context.Context, api redfishAPI.RedfishAPI, systemID string) (string, string, error) {
	managerID, err := GetManagerID(ctx, api, systemID)
	if err != nil {
		return "", "", err
	}

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

// ScreenRedfishError provides a detailed error message for end user consumption by inspecting all Redfish client
// responses and errors.
func ScreenRedfishError(httpResp *http.Response, clientErr error) error {
	if httpResp == nil {
		return ErrRedfishClient{
			Message: "HTTP request failed. Redfish may be temporarily unavailable. Please try again.",
		}
	}

	// NOTE(drewwalters96): The error, clientErr, may not be nil even though the request was successful. The HTTP
	// status code is the most reliable way to determine the result of a Redfish request using the go-redfish
	// library. The Redfish client uses HTTP codes 200 and 204 to indicate success.
	var finalError ErrRedfishClient
	switch httpResp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		finalError = ErrRedfishClient{Message: "System not found. Correct the system name and try again."}
	case http.StatusBadRequest:
		finalError = ErrRedfishClient{Message: "Invalid request. Verify the system name and try again."}
	case http.StatusMethodNotAllowed:
		finalError = ErrRedfishClient{
			Message: fmt.Sprintf("%s. BMC returned status '%s'.",
				"This operation is likely unsupported by the BMC Redfish version, or the BMC is busy",
				httpResp.Status),
		}
	default:
		finalError = ErrRedfishClient{Message: fmt.Sprintf("BMC responded '%s'.", httpResp.Status)}
		log.Debugf("BMC responded '%s'. Attempting to unmarshal the raw BMC error response.",
			httpResp.Status)
	}

	// NOTE(drewwalters96) Check for error messages with extended information, as they often accompany more
	// general error descriptions provided in "clientErr". Since there can be multiple errors, wrap them in a
	// single error.
	var redfishErr redfishClient.RedfishError
	var additionalInfo string

	oAPIErr, ok := clientErr.(redfishClient.GenericOpenAPIError)
	if ok {
		if err := json.Unmarshal(oAPIErr.Body(), &redfishErr); err == nil {
			additionalInfo = ""
			for _, extendedInfo := range redfishErr.Error.MessageExtendedInfo {
				additionalInfo = fmt.Sprintf("%s %s %s", additionalInfo, extendedInfo.Message,
					extendedInfo.Resolution)
			}
		} else {
			log.Debugf("Unable to unmarshal the raw BMC error response. %s", err.Error())
		}
	}

	if strings.TrimSpace(additionalInfo) != "" {
		finalError.Message = fmt.Sprintf("%s %s", finalError.Message, additionalInfo)
	} else if redfishErr.Error.Message != "" {
		// Provide the general error message when there were no messages containing extended information.
		finalError.Message = fmt.Sprintf("%s %s", finalError.Message, redfishErr.Error.Message)
	}

	return finalError
}

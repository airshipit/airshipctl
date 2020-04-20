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
	"net/http"
	"net/url"
	"strings"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

// URLSchemeSeparator holds the separator for URL scheme
// Example: git+ssh
const (
	URLSchemeSeparator = "+"
)

// GetManagerID retrieves the manager ID for a redfish system.
func GetManagerID(ctx context.Context, api redfishAPI.RedfishAPI, systemID string) (string, error) {
	system, _, err := api.GetSystem(ctx, systemID)
	if err != nil {
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

// ScreenRedfishError provides detailed error checking on a Redfish client response.
func ScreenRedfishError(httpResp *http.Response, clientErr error) error {
	if httpResp == nil {
		return ErrRedfishClient{Message: "HTTP request failed. Please try again."}
	}

	// NOTE(drewwalters96): clientErr may not be nil even though the request was successful. The HTTP status code
	// has to be verified for success on each request. The Redfish client uses HTTP codes 200 and 204 to indicate
	// success.
	if httpResp.StatusCode >= http.StatusOK && httpResp.StatusCode <= http.StatusNoContent {
		// This range of status codes indicate success
		return nil
	}

	if clientErr == nil {
		return ErrRedfishClient{Message: http.StatusText(httpResp.StatusCode)}
	}

	oAPIErr, ok := clientErr.(redfishClient.GenericOpenAPIError)
	if !ok {
		return ErrRedfishClient{Message: "Unable to decode client error."}
	}

	var resp redfishClient.RedfishError
	if err := json.Unmarshal(oAPIErr.Body(), &resp); err != nil {
		// No JSON response included; use generic error text.
		return ErrRedfishClient{Message: err.Error()}
	}

	return ErrRedfishClient{Message: resp.Error.Message}
}

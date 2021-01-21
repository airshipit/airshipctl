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
	"time"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/log"
)

// URLSchemeSeparator holds the separator for URL scheme
// Example: git+ssh
const (
	redfishURLSchemeSeparator = "+"
)

func processExtendedInfo(extendedInfo map[string]interface{}) (string, error) {
	message, ok := extendedInfo["Message"]
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo.Message"}
	}

	messageContent, ok := message.(string)
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo.Message"}
	}

	// Resolution may be omitted in some responses
	if resolution, ok := extendedInfo["Resolution"]; ok {
		return fmt.Sprintf("%s %s", messageContent, resolution), nil
	}

	return messageContent, nil
}

// DecodeRawError decodes a raw Redfish HTTP response and retrieves the extended information and available resolutions
// returned by the BMC.
func DecodeRawError(rawResponse []byte) (string, error) {
	// Unmarshal raw Redfish response as arbitrary JSON map
	var arbitraryJSON map[string]interface{}
	if err := json.Unmarshal(rawResponse, &arbitraryJSON); err != nil {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	errObject, ok := arbitraryJSON["error"]
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	errContent, ok := errObject.(map[string]interface{})
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error"}
	}

	extendedInfoContent, ok := errContent["@Message.ExtendedInfo"]
	if !ok {
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
	}

	// NOTE(drewwalters96): The official specification dictates that "@Message.ExtendedInfo" should be a JSON array;
	// however, some BMCs have returned a single JSON dictionary. Handle both types here.
	switch extendedInfo := extendedInfoContent.(type) {
	case []interface{}:
		if len(extendedInfo) == 0 {
			return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
		}

		var errorMessage string
		for _, info := range extendedInfo {
			infoContent, ok := info.(map[string]interface{})
			if !ok {
				return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
			}

			if _, ok := infoContent["Message"]; !ok {
				return errContent["message"].(string), nil
			}

			message, err := processExtendedInfo(infoContent)
			if err != nil {
				return "", err
			}

			errorMessage = fmt.Sprintf("%s\n%s", message, errorMessage)
		}

		return errorMessage, nil
	case map[string]interface{}:
		return processExtendedInfo(extendedInfo)
	default:
		return "", ErrUnrecognizedRedfishResponse{Key: "error.@Message.ExtendedInfo"}
	}
}

// GetManagerID retrieves the manager ID for a redfish system.
func GetManagerID(ctx context.Context, api redfishAPI.RedfishAPI, systemID string) (string, error) {
	system, httpResp, err := api.GetSystem(ctx, systemID)
	if err = ScreenRedfishError(httpResp, err); err != nil {
		log.Debugf("Unable to find manager for node '%s'.", systemID)
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
	log.Debug("Searching for compatible media types.")
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
				log.Debugf("Found virtual media type '%s' with ID '%s' on manager '%s'.", mediaType,
					mediaID, managerID)
				return mediaID, mediaType, nil
			}
		}
	}

	return "", "", ErrRedfishClient{Message: fmt.Sprintf("Manager '%s' does not have virtual media type CD or DVD.",
		managerID)}
}

// ScreenRedfishError provides a detailed error message for end user consumption by inspecting all Redfish client
// responses and errors.
func ScreenRedfishError(httpResp *http.Response, clientErr error) error {
	if httpResp == nil {
		m := "HTTP request failed. Redfish may be temporarily unavailable. Please try again."
		if clientErr != nil {
			m += fmt.Sprintf(" original error %s", clientErr.Error())
		}
		return ErrRedfishClient{
			Message: m,
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
		finalError = ErrRedfishClient{Message: httpResp.Status}
	}

	// Retrieve the raw HTTP response body
	oAPIErr, ok := clientErr.(redfishClient.GenericOpenAPIError)
	if !ok {
		log.Debug("Unable to decode BMC response.")
	}

	// Attempt to decode the BMC response from the raw HTTP response
	if bmcResponse, err := DecodeRawError(oAPIErr.Body()); err == nil {
		finalError.Message = fmt.Sprintf("%s\nBMC responded: '%s'", finalError.Message, bmcResponse)
	} else {
		log.Debugf("Unable to decode BMC response. %q", err)
	}

	return finalError
}

// SetAuth allows to set username and password to given context so that redfish client can
// authenticate against redfish server
func SetAuth(ctx context.Context, username string, password string) context.Context {
	authValue := redfishClient.BasicAuth{UserName: username, Password: password}
	if ctx.Value(redfishClient.ContextBasicAuth) == authValue {
		return ctx
	}
	return context.WithValue(
		ctx,
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: username, Password: password},
	)
}

func getManagerID(ctx context.Context, api redfishAPI.RedfishAPI, systemID string) (string, error) {
	system, _, err := api.GetSystem(ctx, systemID)
	if err != nil {
		return "", err
	}

	return GetResourceIDFromURL(system.Links.ManagedBy[0].OdataId), nil
}

func getBasePath(redfishURL string) (string, error) {
	parsedURL, err := url.Parse(redfishURL)
	if err != nil {
		return "", ErrRedfishClient{Message: fmt.Sprintf("Redfish URL malformed %s", err.Error())}
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
	schemeSplit := strings.Split(parsedURL.Scheme, redfishURLSchemeSeparator)
	if len(schemeSplit) > 1 {
		baseURL = fmt.Sprintf("%s://%s", schemeSplit[len(schemeSplit)-1], parsedURL.Host)
	}

	return baseURL, nil
}

func (c Client) waitForPowerState(ctx context.Context, desiredState redfishClient.PowerState) error {
	log.Debugf("Waiting for node '%s' to reach power state '%s'.", c.nodeID, desiredState)

	for retry := 0; retry <= c.systemActionRetries; retry++ {
		system, httpResp, err := c.RedfishAPI.GetSystem(ctx, c.NodeID())
		if err = ScreenRedfishError(httpResp, err); err != nil {
			return err
		}

		if system.PowerState == desiredState {
			log.Debugf("Node '%s' reached power state '%s'.", c.nodeID, desiredState)
			return nil
		}

		c.Sleep(time.Duration(c.systemRebootDelay) * time.Second)
	}

	return ErrOperationRetriesExceeded{
		What:    fmt.Sprintf("reach desired power state %s", desiredState),
		Retries: c.systemActionRetries,
	}
}

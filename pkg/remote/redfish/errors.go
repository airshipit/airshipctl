package redfish

import (
	"encoding/json"
	"fmt"
	"net/http"

	redfishClient "opendev.org/airship/go-redfish/client"

	aerror "opendev.org/airship/airshipctl/pkg/errors"
)

// ErrRedfishClient describes an error encountered by the go-redfish client.
type ErrRedfishClient struct {
	aerror.AirshipError
	Message string
}

func (e ErrRedfishClient) Error() string {
	return fmt.Sprintf("redfish client encountered an error: %s", e.Message)
}

// ErrRedfishMissingConfig describes an error encountered due to a missing configuration option.
type ErrRedfishMissingConfig struct {
	What string
}

func (e ErrRedfishMissingConfig) Error() string {
	return "missing configuration: " + e.What
}

// ErrOperationRetriesExceeded raised if number of operation retries exceeded
type ErrOperationRetriesExceeded struct {
}

func (e ErrOperationRetriesExceeded) Error() string {
	return "maximum retries exceeded"
}

// ScreenRedfishError provides detailed error checking on a Redfish client response.
func ScreenRedfishError(httpResp *http.Response, clientErr error) error {
	// NOTE(drewwalters96): clientErr may not be nil even though the request was successful. The HTTP status code
	// has to be verified for success on each request. The Redfish client uses HTTP codes 200 and 204 to indicate
	// success.
	if httpResp != nil && httpResp.StatusCode != 200 && httpResp.StatusCode != 204 {
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
	} else if httpResp == nil {
		return ErrRedfishClient{Message: "HTTP request failed. Please try again."}
	}

	return nil
}

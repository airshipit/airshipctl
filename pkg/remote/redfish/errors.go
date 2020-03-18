package redfish

import (
	"fmt"
	"net/http"

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

// Redfish client return error if response has no body.
// In this case this function further checks the
// HTTP response code.
func ScreenRedfishError(httpResp *http.Response, err error) error {
	if err != nil && httpResp != nil {
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 400 {
			return ErrRedfishClient{Message: err.Error()}
		}
	} else if err != nil {
		return ErrRedfishClient{Message: err.Error()}
	}
	return nil
}

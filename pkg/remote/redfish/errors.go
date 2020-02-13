package redfish

import (
	"fmt"
	"net/http"

	aerror "opendev.org/airship/airshipctl/pkg/errors"
)

type RedfishClientError struct {
	aerror.AirshipError
}

func NewRedfishClientErrorf(format string, v ...interface{}) error {
	e := &RedfishClientError{}
	e.Message = fmt.Sprintf(format, v...)
	return e
}

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
			return NewRedfishClientErrorf("%s", err.Error())
		}
	} else if err != nil {
		return NewRedfishClientErrorf("%s", err.Error())
	}
	return nil
}

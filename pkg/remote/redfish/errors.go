package redfish

import (
	"fmt"

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

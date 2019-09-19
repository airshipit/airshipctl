package remote

import (
	"fmt"

	aerror "opendev.org/airship/airshipctl/pkg/error"
)

type RemoteDirectError struct {
	aerror.AirshipError
}

func NewRemoteDirectErrorf(format string, v ...interface{}) error {
	e := &RemoteDirectError{}
	e.Message = fmt.Sprintf(format, v...)
	return e
}

package remote

import (
	"fmt"

	aerror "opendev.org/airship/airshipctl/pkg/errors"
)

type GenericError struct {
	aerror.AirshipError
}

func NewRemoteDirectErrorf(format string, v ...interface{}) error {
	e := &GenericError{}
	e.Message = fmt.Sprintf(format, v...)
	return e
}

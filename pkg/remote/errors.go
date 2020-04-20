/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package remote

import (
	"fmt"

	aerror "opendev.org/airship/airshipctl/pkg/errors"
)

// TODO: This need to be refactored to match the error format used elsewhere in airshipctl
// Usage of this error should be deprecated as it doesn't provide meaningful feedback to the user.

// GenericError provides general feedback about an error that occurred in a remote operation
type GenericError struct {
	aerror.AirshipError
}

// NewRemoteDirectErrorf retruns formatted remote direct errors
func NewRemoteDirectErrorf(format string, v ...interface{}) error {
	e := &GenericError{}
	e.Message = fmt.Sprintf(format, v...)
	return e
}

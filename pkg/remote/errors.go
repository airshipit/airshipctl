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
)

// TODO: This need to be refactored to match the error format used elsewhere in airshipctl
// Usage of this error should be deprecated as it doesn't provide meaningful feedback to the user.

// GenericError provides general feedback about an error that occurred in a remote operation
type GenericError struct {
	Message string
}

// NewRemoteDirectErrorf retruns formatted remote direct errors
func NewRemoteDirectErrorf(format string, v ...interface{}) error {
	return &GenericError{Message: fmt.Sprintf(format, v...)}
}

func (e GenericError) Error() string {
	return e.Message
}

// ErrUnknownManagementType is an error that indicates the remote type specified in the airshipctl management
// configuration (e.g. redfish, redfish-dell) is not supported.
type ErrUnknownManagementType struct {
	Type string
}

func (e ErrUnknownManagementType) Error() string {
	return fmt.Sprintf("unknown management type: %s", e.Type)
}

// ErrMissingBootstrapInfoOption is an error that indicates a bootstrap option is missing in the airshipctl
// bootstrapInfo configuration.
type ErrMissingBootstrapInfoOption struct {
	What string
}

func (e ErrMissingBootstrapInfoOption) Error() string {
	return fmt.Sprintf("missing bootstrapInfo option: %s", e.What)
}

// ErrNoHostsFound is an error that indicates that no hosts matched the selection criteria passed to a manager.
type ErrNoHostsFound struct{}

func (e ErrNoHostsFound) Error() string {
	return "no hosts selected"
}

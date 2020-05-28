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

package redfish

import (
	"fmt"
)

// ErrRedfishClient describes an error encountered by the go-redfish client.
type ErrRedfishClient struct {
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
	What    string
	Retries int
}

func (e ErrOperationRetriesExceeded) Error() string {
	return fmt.Sprintf("Unable to %s. Maximum retries (%d) exceeded.", e.What, e.Retries)
}

// ErrUnrecognizedRedfishResponse is a debug error that describes unexpected formats in a Redfish error response.
type ErrUnrecognizedRedfishResponse struct {
	Key string
}

func (e ErrUnrecognizedRedfishResponse) Error() string {
	return fmt.Sprintf("Unable to decode Redfish response. Key '%s' is missing or has unknown format.", e.Key)
}

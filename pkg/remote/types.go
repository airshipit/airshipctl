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

package remote

import (
	"context"
)

// Client is a set of functions that clients created for out-of-band power management and control should implement. The
// functions within client are used by power management commands and remote direct functionality.
type Client interface {
	RebootSystem(context.Context, string) error

	// TODO(drewwalters96): Should this be a string forever? We may want to define our own custom type, as the
	// string format will be client dependent when we add new clients.
	SystemPowerStatus(context.Context, string) (string, error)
	EphemeralNodeID() string

	// TODO(drewwalters96): This function may be too tightly coupled to remoteDirect operations. This could probably
	// be combined with SetVirtualMedia.
	SetEphemeralBootSourceByType(context.Context) error

	// TODO(drewwalters96): This function is tightly coupled to Redfish. It should be combined with the
	// SetBootSource operation and removed from the client interface.
	SetVirtualMedia(context.Context, string) error
}

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

package ifc

import (
	"context"

	"opendev.org/airship/airshipctl/pkg/remote/power"
)

// Client is a set of functions that clients created for out-of-band power management and control should implement. The
// functions within client are used by power management commands and remote direct functionality.
type Client interface {
	EjectVirtualMedia(context.Context) error
	NodeID() string
	NodeName() string
	RebootSystem(context.Context) error
	SetBootSourceByType(context.Context) error
	SystemPowerOff(context.Context) error
	SystemPowerOn(context.Context) error
	SystemPowerStatus(context.Context) (power.Status, error)
	RemoteDirect(context.Context, string) error

	// TODO(drewwalters96): This function is tightly coupled to Redfish. It should be combined with the
	// SetBootSource operation and removed from the client interface.
	SetVirtualMedia(context.Context, string) error
}

// ClientFactory is a function to be used
type ClientFactory func(name string,
	redfishURL string,
	insecure bool, useProxy bool,
	username string, password string,
	systemActionRetries int, systemRebootDelay int) (Client, error)

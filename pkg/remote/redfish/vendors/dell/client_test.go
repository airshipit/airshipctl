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

package dell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ephemeralNodeID = "System.Embedded.1"
	isoPath         = "https://localhost:8080/debian.iso"
	redfishURL      = "redfish+https://localhost/Systems/System.Embedded.1"
)

func TestNewClient(t *testing.T) {
	// NOTE(drewwalters96): The Dell client implementation of this method simply creates the standard Redfish
	// client. This test verifies that the Dell client creates and stores an instance of the standard client.

	// Create the Dell client
	_, _, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "username", "password")
	assert.NoError(t, err)
}

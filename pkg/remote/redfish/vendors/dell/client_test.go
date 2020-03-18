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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"
)

const (
	ephemeralNodeID = "System.Embedded.1"
	isoPath         = "https://localhost:8080/debian.iso"
	redfishURL      = "redfish+https://localhost/Systems/System.Embedded.1"
)

func TestNewClient(t *testing.T) {
	_, _, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "username", "password")
	assert.NoError(t, err)
}

func TestSetEphemeralBootSourceByTypeGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.EphemeralNodeID()).Times(1).Return(redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx)
	assert.Error(t, err)
}

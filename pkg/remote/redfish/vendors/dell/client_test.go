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

package dell

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"
)

const (
	redfishURL          = "redfish+https://localhost/Systems/System.Embedded.1"
	systemActionRetries = 0
	systemRebootDelay   = 0
)

func TestNewClient(t *testing.T) {
	_, err := newClient(redfishURL, false, false, "username", "password", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
}

func TestNewClientInterface(t *testing.T) {
	c, err := ClientFactory(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewClientDefaultValues(t *testing.T) {
	sysActRetr := 222
	sysRebDel := 555
	c, err := newClient(redfishURL, false, false, "", "", sysActRetr, sysRebDel)
	assert.Equal(t, c.SystemActionRetries(), sysActRetr)
	assert.Equal(t, c.SystemRebootDelay(), sysRebDel)
	assert.NoError(t, err)
}
func TestSetBootSourceByTypeGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := newClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
	ctx := context.Background()

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.NodeID()).Times(1).Return(redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	err = client.SetBootSourceByType(ctx)
	assert.Error(t, err)
}

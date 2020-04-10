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
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"

	testutil "opendev.org/airship/airshipctl/testutil/redfishutils/helpers"
)

const (
	ephemeralNodeID = "ephemeral-node-id"
	isoPath         = "https://localhost:8080/debian.iso"
	redfishURL      = "https://localhost:1234"
)

func TestNewClient(t *testing.T) {
	_, _, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)
}

func TestNewClientAuth(t *testing.T) {
	ctx, _, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "username", "password")
	assert.NoError(t, err)

	cAuth := ctx.Value(redfishClient.ContextBasicAuth)
	auth := redfishClient.BasicAuth{UserName: "username", Password: "password"}
	assert.Equal(t, cAuth, auth)
}

func TestNewClientEmptyRedfishURL(t *testing.T) {
	// Redfish URL cannot be empty when creating a client.
	_, _, err := NewClient(ephemeralNodeID, isoPath, "", false, false, "", "")
	assert.Error(t, err)
}

func TestRebootSystem(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	// Mock redfish shutdown and status requests
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	httpResp := &http.Response{StatusCode: 200}
	m.On("ResetSystem", ctx, ephemeralNodeID, resetReq).Times(1).Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, ephemeralNodeID).Times(1).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF}, httpResp, nil)

	// Mock redfish startup and status requests
	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, ephemeralNodeID, resetReq).Times(1).Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, ephemeralNodeID).Times(1).
		Return(redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON}, httpResp, nil)

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.RebootSystem(ctx, ephemeralNodeID)
	assert.NoError(t, err)
}

func TestRebootSystemShutdownError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	// Mock redfish shutdown request for failure
	m.On("ResetSystem", ctx, ephemeralNodeID, resetReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.RebootSystem(ctx, ephemeralNodeID)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRebootSystemStartupError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	// Mock redfish shutdown request
	systemID := ephemeralNodeID
	m.On("ResetSystem", ctx, systemID, resetReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, systemID).Times(1).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF},
		&http.Response{StatusCode: 200}, nil)

	resetOnReq := redfishClient.ResetRequestBody{}
	resetOnReq.ResetType = redfishClient.RESETTYPE_ON

	// Mock redfish startup request for failure
	m.On("ResetSystem", ctx, systemID, resetOnReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.RebootSystem(ctx, systemID)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRebootSystemTimeout(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(ephemeralNodeID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), "numRetries", 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	systemID := ephemeralNodeID
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, systemID).
		Return(redfishClient.ComputerSystem{}, &http.Response{StatusCode: 200}, nil)

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.RebootSystem(ctx, systemID)
	assert.Equal(t, ErrOperationRetriesExceeded{}, err)
}

func TestSetEphemeralBootSourceByTypeGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.ephemeralNodeID).Times(1).Return(redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx)
	assert.Error(t, err)
}

func TestSetEphemeralBootSourceByTypeSetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.ephemeralNodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	m.On("SetSystem", ctx, client.ephemeralNodeID, mock.Anything).Times(1).Return(
		redfishClient.ComputerSystem{}, &http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx)
	assert.Error(t, err)
}

func TestSetEphemeralBootSourceByTypeBootSourceUnavailable(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	invalidSystem := testutil.GetTestSystem()
	invalidSystem.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	}

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.ephemeralNodeID).Return(invalidSystem, nil, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSetVirtualMediaGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.ephemeralNodeID).Times(1).Return(redfishClient.ComputerSystem{},
		nil, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetVirtualMedia(ctx, client.isoPath)
	assert.Error(t, err)
}

func TestSetVirtualMediaInsertVirtualMediaError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := ephemeralNodeID
	ctx, client, err := NewClient(systemID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.ephemeralNodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	m.On("InsertVirtualMedia", context.Background(), testutil.ManagerID, "Cd", mock.Anything).Return(
		redfishClient.RedfishError{}, &http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetVirtualMedia(ctx, client.isoPath)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

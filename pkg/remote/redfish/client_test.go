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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/remote/power"
	testutil "opendev.org/airship/airshipctl/testutil/redfishutils/helpers"
)

const (
	nodeID     = "System.Embedded.1"
	isoPath    = "https://localhost:8080/debian.iso"
	redfishURL = "redfish+https://localhost:2224/Systems/System.Embedded.1"
)

func TestNewClient(t *testing.T) {
	_, _, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)
}

func TestNewClientMissingSystemID(t *testing.T) {
	badURL := "redfish+https://localhost:2224"

	_, _, err := NewClient(badURL, false, false, "", "")
	_, ok := err.(ErrRedfishMissingConfig)
	assert.True(t, ok)
}

func TestNewClientNoRedfishMarking(t *testing.T) {
	url := "https://localhost:2224/Systems/System.Embedded.1"

	_, _, err := NewClient(url, false, false, "", "")
	assert.NoError(t, err)
}

func TestNewClientAuth(t *testing.T) {
	ctx, _, err := NewClient(redfishURL, false, false, "username", "password")
	assert.NoError(t, err)

	cAuth := ctx.Value(redfishClient.ContextBasicAuth)
	auth := redfishClient.BasicAuth{UserName: "username", Password: "password"}
	assert.Equal(t, cAuth, auth)
}

func TestNewClientEmptyRedfishURL(t *testing.T) {
	// Redfish URL cannot be empty when creating a client.
	_, _, err := NewClient("", false, false, "", "")
	assert.Error(t, err)
}

func TestEjectVirtualMedia(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Normal retries are 30. Limit them here for test time.
	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 2)

	// Mark CD and DVD test media as inserted
	inserted := true
	testMediaCD := testutil.GetVirtualMedia([]string{"CD"})
	testMediaCD.Inserted = &inserted

	testMediaDVD := testutil.GetVirtualMedia([]string{"DVD"})
	testMediaDVD.Inserted = &inserted

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd", "DVD", "Floppy"}), httpResp, nil)

	// Eject CD
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMediaCD, httpResp, nil)
	m.On("EjectVirtualMedia", ctx, testutil.ManagerID, "Cd", mock.Anything).Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"Cd"}), httpResp, nil)

	// Eject DVD and simulate two retries
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "DVD").Times(1).
		Return(testMediaDVD, httpResp, nil)
	m.On("EjectVirtualMedia", ctx, testutil.ManagerID, "DVD", mock.Anything).Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "DVD").Times(1).
		Return(testMediaDVD, httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "DVD").Times(1).
		Return(testutil.GetVirtualMedia([]string{"DVD"}), httpResp, nil)

	// Floppy is not inserted, so it is not ejected
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Floppy").Times(1).
		Return(testutil.GetVirtualMedia([]string{"Floppy"}), httpResp, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.EjectVirtualMedia(ctx)
	assert.NoError(t, err)
}
func TestEjectVirtualMediaRetriesExceeded(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	// Mark test media as inserted
	inserted := true
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.Inserted = &inserted

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMedia, httpResp, nil)

	// Verify retry logic
	m.On("EjectVirtualMedia", ctx, testutil.ManagerID, "Cd", mock.Anything).Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)

	// Media still inserted on retry. Since retries are 1, this causes failure.
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMedia, httpResp, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.EjectVirtualMedia(ctx)
	_, ok := err.(ErrOperationRetriesExceeded)
	assert.True(t, ok)
}
func TestRebootSystem(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Mock redfish shutdown and status requests
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	httpResp := &http.Response{StatusCode: 200}
	m.On("ResetSystem", ctx, client.nodeID, resetReq).Times(1).Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, client.nodeID).Times(1).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF}, httpResp, nil)

	// Mock redfish startup and status requests
	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, client.nodeID, resetReq).Times(1).Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, client.nodeID).Times(1).
		Return(redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON}, httpResp, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.RebootSystem(ctx)
	assert.NoError(t, err)
}

func TestRebootSystemShutdownError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	// Mock redfish shutdown request for failure
	m.On("ResetSystem", ctx, client.nodeID, resetReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.RebootSystem(ctx)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRebootSystemStartupError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	// Mock redfish shutdown request
	m.On("ResetSystem", ctx, client.nodeID, resetReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, client.nodeID).Times(1).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF},
		&http.Response{StatusCode: 200}, nil)

	resetOnReq := redfishClient.ResetRequestBody{}
	resetOnReq.ResetType = redfishClient.RESETTYPE_ON

	// Mock redfish startup request for failure
	m.On("ResetSystem", ctx, client.nodeID, resetOnReq).Times(1).Return(redfishClient.RedfishError{},
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.RebootSystem(ctx)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRebootSystemTimeout(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("ResetSystem", ctx, client.nodeID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, client.nodeID).
		Return(redfishClient.ComputerSystem{}, &http.Response{StatusCode: 200}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.RebootSystem(ctx)
	assert.Error(t, err)
}

func TestSetBootSourceByTypeGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.NodeID()).Times(1).Return(redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetBootSourceByType(ctx)
	assert.Error(t, err)
}

func TestSetBootSourceByTypeSetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	m.On("SetSystem", ctx, client.nodeID, mock.Anything).Times(1).Return(
		redfishClient.ComputerSystem{}, &http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetBootSourceByType(ctx)
	assert.Error(t, err)
}

func TestSetBootSourceByTypeBootSourceUnavailable(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	invalidSystem := testutil.GetTestSystem()
	invalidSystem.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	}

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(invalidSystem, httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetBootSourceByType(ctx)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSetVirtualMediaEjectExistingMedia(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Normal retries are 30. Limit them here for test time.
	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	// Mark test media as inserted
	inserted := true
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.Inserted = &inserted

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)

	// Eject Media calls
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMedia, httpResp, nil)
	m.On("EjectVirtualMedia", ctx, testutil.ManagerID, "Cd", mock.Anything).Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Insert media calls
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	m.On("InsertVirtualMedia", ctx, testutil.ManagerID, "Cd", mock.Anything).Return(
		redfishClient.RedfishError{}, httpResp, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetVirtualMedia(ctx, isoPath)
	assert.NoError(t, err)
}

func TestSetVirtualMediaEjectExistingMediaFailure(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Normal retries are 30. Limit them here for test time.
	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	// Mark test media as inserted
	inserted := true
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.Inserted = &inserted

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)

	// Eject Media calls
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMedia, httpResp, nil)
	m.On("EjectVirtualMedia", ctx, testutil.ManagerID, "Cd", mock.Anything).Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testMedia, httpResp, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetVirtualMedia(ctx, isoPath)
	assert.Error(t, err)
}
func TestSetVirtualMediaGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	// Mock redfish get system request
	m.On("GetSystem", ctx, client.nodeID).Times(1).Return(redfishClient.ComputerSystem{},
		nil, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetVirtualMedia(ctx, isoPath)
	assert.Error(t, err)
}

func TestSetVirtualMediaInsertVirtualMediaError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	client.nodeID = nodeID

	httpResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, client.nodeID).Return(testutil.GetTestSystem(), httpResp, nil)
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Insert media calls
	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil)
	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	m.On("InsertVirtualMedia", context.Background(), testutil.ManagerID, "Cd", mock.Anything).Return(
		redfishClient.RedfishError{}, &http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.RedfishAPI = m
	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SetVirtualMedia(ctx, isoPath)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSystemPowerOff(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	m.On("ResetSystem", ctx, client.nodeID, mock.Anything).Return(
		redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON},
		&http.Response{StatusCode: 200}, nil).Times(1)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF},
		&http.Response{StatusCode: 200}, nil).Times(1)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SystemPowerOff(ctx)
	assert.NoError(t, err)
}

func TestSystemPowerOffResetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	m.On("ResetSystem", ctx, client.nodeID, mock.Anything).Return(
		redfishClient.RedfishError{},
		&http.Response{StatusCode: 500}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SystemPowerOff(ctx)
	assert.Error(t, err)
}

func TestSystemPowerOn(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	m.On("ResetSystem", ctx, client.nodeID, mock.Anything).Return(
		redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF},
		&http.Response{StatusCode: 200}, nil).Times(1)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON},
		&http.Response{StatusCode: 200}, nil).Times(1)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SystemPowerOn(ctx)
	assert.NoError(t, err)
}

func TestSystemPowerOnResetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)

	m.On("ResetSystem", ctx, client.nodeID, mock.Anything).Return(
		redfishClient.RedfishError{},
		&http.Response{StatusCode: 500}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.SystemPowerOn(ctx)
	assert.Error(t, err)
}

func TestSystemPowerStatusUnknown(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	var unknownState redfishClient.PowerState = "unknown"
	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: unknownState},
		&http.Response{StatusCode: 200},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusUnknown, status)
}

func TestSystemPowerStatusOn(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON},
		&http.Response{StatusCode: 200},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusOn, status)
}

func TestSystemPowerStatusOff(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF},
		&http.Response{StatusCode: 200},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusOff, status)
}

func TestSystemPowerStatusPoweringOn(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_POWERING_ON},
		&http.Response{StatusCode: 200},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusPoweringOn, status)
}

func TestSystemPowerStatusPoweringOff(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_POWERING_OFF},
		&http.Response{StatusCode: 200},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusPoweringOff, status)
}

func TestSystemPowerStatusGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(redfishURL, false, false, "", "")
	require.NoError(t, err)

	client.nodeID = nodeID

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500},
		redfishClient.GenericOpenAPIError{})

	client.RedfishAPI = m

	_, err = client.SystemPowerStatus(ctx)
	assert.Error(t, err)
}

func TestWaitForPowerStateGetSystemFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{}, &http.Response{StatusCode: 500}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF)
	assert.Error(t, err)
}

func TestWaitForPowerStateNoRetries(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_OFF,
		}, &http.Response{StatusCode: 200}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF)
	assert.NoError(t, err)
}

func TestWaitForPowerStateWithRetries(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_ON,
		}, &http.Response{StatusCode: 200}, nil).Times(1)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_OFF,
		}, &http.Response{StatusCode: 200}, nil).Times(1)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF)
	assert.NoError(t, err)
}

func TestWaitForPowerStateRetriesExceeded(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_ON,
		}, &http.Response{StatusCode: 200}, nil)

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_ON,
		}, &http.Response{StatusCode: 200}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_OFF)
	assert.Error(t, err)
}

func TestWaitForPowerStateDifferentPowerState(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, client, err := NewClient(redfishURL, false, false, "", "")
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), ctxKeyNumRetries, 1)
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_ON

	m.On("GetSystem", ctx, client.nodeID).Return(
		redfishClient.ComputerSystem{
			PowerState: redfishClient.POWERSTATE_ON,
		}, &http.Response{StatusCode: 200}, nil)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_ON)
	assert.NoError(t, err)
}

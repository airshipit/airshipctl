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
	"github.com/stretchr/testify/require"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/remote/power"
	testutil "opendev.org/airship/airshipctl/testutil/redfishutils/helpers"
)

const (
	nodeID              = "System.Embedded.1"
	isoPath             = "http://localhost:8099/ubuntu-focal.iso"
	redfishURL          = "redfish+https://localhost:2224/Systems/System.Embedded.1"
	systemActionRetries = 1
	systemRebootDelay   = 0
)

func TestNewClient(t *testing.T) {
	c, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewClientInterface(t *testing.T) {
	c, err := ClientFactory(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestNewClientDefaultValues(t *testing.T) {
	sysActRetr := 111
	sysRebDel := 999
	c, err := NewClient(redfishURL, false, false, "", "", sysActRetr, sysRebDel)
	assert.Equal(t, c.systemActionRetries, sysActRetr)
	assert.Equal(t, c.systemRebootDelay, sysRebDel)
	assert.NoError(t, err)
}

func TestNewClientMissingSystemID(t *testing.T) {
	badURL := "redfish+https://localhost:2224"

	_, err := NewClient(badURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	_, ok := err.(ErrRedfishMissingConfig)
	assert.True(t, ok)
}

func TestNewClientNoRedfishMarking(t *testing.T) {
	url := "https://localhost:2224/Systems/System.Embedded.1"

	_, err := NewClient(url, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)
}

func TestNewClientEmptyRedfishURL(t *testing.T) {
	// Redfish URL cannot be empty when creating a client.
	_, err := NewClient("", false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.Error(t, err)
}
func TestEjectVirtualMedia(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries+1, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	client.RedfishAPI = m

	ctx := SetAuth(context.Background(), "", "")

	// Mark CD and DVD test media as inserted
	inserted := true
	testMediaCD := testutil.GetVirtualMedia([]string{"CD"})
	testMediaCD.SetInserted(inserted)

	testMediaDVD := testutil.GetVirtualMedia([]string{"DVD"})
	testMediaDVD.SetInserted(inserted)

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(), httpResp, nil, 1)

	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd", "DVD", "Floppy"}), httpResp, nil, 1)
	// Eject CD
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID,
		"Cd", testMediaCD, httpResp, nil)
	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, nil)

	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"Cd"}), httpResp, nil)

	// Eject DVD and simulate two retries
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID,
		"DVD", testMediaDVD, httpResp, nil)
	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "DVD",
		redfishClient.RedfishError{}, httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID,
		"DVD", testMediaDVD, httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID,
		"DVD", testutil.GetVirtualMedia([]string{"DVD"}), httpResp, nil)

	// Floppy is not inserted, so it is not ejected
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID,
		"Floppy", testutil.GetVirtualMedia([]string{"Floppy"}), httpResp, nil)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID

	ctx := SetAuth(context.Background(), "", "")

	// Mark test media as inserted
	inserted := true
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.SetInserted(inserted)

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(), httpResp, nil, 1)

	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMedia, httpResp, nil)

	// Verify retry logic
	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, nil)

	// Media still inserted on retry. Since retries are 1, this causes failure.
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMedia, httpResp, nil)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	// Mock redfish shutdown and status requests
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)
	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetReq, redfishClient.RedfishError{}, httpResp, nil)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem, httpResp, nil, 1)

	// Mock redfish startup and status requests
	resetReq.SetResetType(redfishClient.RESETTYPE_ON)
	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetReq, redfishClient.RedfishError{}, httpResp, nil)
	computerSystem = redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem, httpResp, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	// Mock redfish shutdown request for failure
	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetReq, redfishClient.RedfishError{},
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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	// Mock redfish shutdown request

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetReq, redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

	resetOnReq := redfishClient.ResetRequestBody{}
	resetOnReq.SetResetType(redfishClient.RESETTYPE_ON)

	// Mock redfish startup request for failure
	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetOnReq, redfishClient.RedfishError{},
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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &resetReq, redfishClient.RedfishError{},
		&http.Response{StatusCode: 200}, nil)

	testutil.MockOnGetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 200}, nil, -1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	// Mock redfish get system request
	testutil.MockOnGetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(),
		httpResp, nil, -1)
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	system := redfishClient.ComputerSystem{}
	boot := redfishClient.NewBoot()
	boot.SetBootSourceOverrideTarget(redfishClient.BOOTSOURCE_CD)
	system.SetBoot(*boot)
	testutil.MockOnSetSystem(ctx, m, client.nodeID, system,
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	client.nodeID = nodeID

	invalidSystem := testutil.GetTestSystem()
	boot := invalidSystem.GetBoot()
	boot.SetBootSourceOverrideTargetRedfishAllowableValues([]redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	})

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, invalidSystem,
		httpResp, nil, 2)
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	system := redfishClient.ComputerSystem{}
	testutil.MockOnSetSystem(ctx, m, client.nodeID, system,
		&http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID

	ctx := SetAuth(context.Background(), "", "")

	// Mark test media as inserted
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.SetInserted(true)

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(),
		httpResp, nil, -1)

	// Eject Media calls
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMedia, httpResp, nil)

	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Insert media calls
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	testutil.MockOnInsertVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	client.nodeID = nodeID

	ctx := SetAuth(context.Background(), "", "")

	// Mark test media as inserted
	inserted := true
	testMedia := testutil.GetVirtualMedia([]string{"CD"})
	testMedia.SetInserted(inserted)

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(),
		httpResp, nil, 1)

	// Eject Media calls
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMedia, httpResp, nil)
	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMedia, httpResp, nil)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	client.nodeID = nodeID

	// Mock redfish get system request
	testutil.MockOnGetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		nil, redfishClient.GenericOpenAPIError{}, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	client.nodeID = nodeID

	httpResp := &http.Response{StatusCode: 200}
	testutil.MockOnGetSystem(ctx, m, client.nodeID, testutil.GetTestSystem(),
		httpResp, nil, 3)
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	// Insert media calls
	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd"}), httpResp, nil, 1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)
	testutil.MockOnInsertVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &redfishClient.ResetRequestBody{},
		redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &redfishClient.ResetRequestBody{},
		redfishClient.RedfishError{}, &http.Response{StatusCode: 500}, nil)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &redfishClient.ResetRequestBody{},
		redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

	computerSystem = redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	testutil.MockOnResetSystem(ctx, m, client.nodeID, &redfishClient.ResetRequestBody{},
		redfishClient.RedfishError{}, &http.Response{StatusCode: 500}, nil)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusUnknown, status)
}

func TestSystemPowerStatusOn(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	client.nodeID = nodeID

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusOn, status)
}

func TestSystemPowerStatusOff(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusOff, status)
}

func TestSystemPowerStatusPoweringOn(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_POWERING_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusPoweringOn, status)
}

func TestSystemPowerStatusPoweringOff(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_POWERING_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	status, err := client.SystemPowerStatus(ctx)
	require.NoError(t, err)

	assert.Equal(t, power.StatusPoweringOff, status)
}

func TestSystemPowerStatusGetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.nodeID = nodeID
	ctx := SetAuth(context.Background(), "", "")
	testutil.MockOnGetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, redfishClient.GenericOpenAPIError{}, 1)

	client.RedfishAPI = m

	_, err = client.SystemPowerStatus(ctx)
	assert.Error(t, err)
}

func TestWaitForPowerStateGetSystemFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		&http.Response{StatusCode: 500}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *computerSystem,
		&http.Response{StatusCode: 200}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID,
		*computerSystem, &http.Response{StatusCode: 200}, nil, 1)

	computerSystem.SetPowerState(redfishClient.POWERSTATE_OFF)
	testutil.MockOnGetSystem(ctx, m, client.nodeID,
		*computerSystem, &http.Response{StatusCode: 200}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_OFF)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID,
		*computerSystem, &http.Response{StatusCode: 200}, nil, 1)

	computerSystem = redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID,
		*computerSystem, &http.Response{StatusCode: 200}, nil, 1)

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

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	assert.NoError(t, err)

	ctx := SetAuth(context.Background(), "", "")
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.SetResetType(redfishClient.RESETTYPE_FORCE_ON)

	computerSystem := redfishClient.NewComputerSystemWithDefaults()
	computerSystem.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID,
		*computerSystem, &http.Response{StatusCode: 200}, nil, 1)

	// Replace normal API client with mocked API client
	client.RedfishAPI = m

	// Mock out the Sleep function so we don't have to wait on it
	client.Sleep = func(_ time.Duration) {}

	err = client.waitForPowerState(ctx, redfishClient.POWERSTATE_ON)
	assert.NoError(t, err)
}

func TestRemoteDirect(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}

	client, err := NewClient(redfishURL, false, false, "", "", systemActionRetries, systemRebootDelay)
	require.NoError(t, err)

	client.RedfishAPI = m

	inserted := true
	testMediaCD := testutil.GetVirtualMedia([]string{"CD"})
	testMediaCD.SetInserted(inserted)
	httpResp := &http.Response{StatusCode: 200}
	system := redfishClient.NewComputerSystemWithDefaults()
	system.SetPowerState(redfishClient.POWERSTATE_OFF)
	links := redfishClient.NewSystemLinksWithDefaults()
	idRef := redfishClient.NewIdRefWithDefaults()
	idRef.SetOdataId(testutil.ManagerID)
	links.SetManagedBy([]redfishClient.IdRef{*idRef})
	system.SetLinks(*links)
	system.SetBoot(redfishClient.Boot{
		BootSourceOverrideTargetRedfishAllowableValues: &[]redfishClient.BootSource{
			redfishClient.BOOTSOURCE_CD,
		},
	})

	ctx := SetAuth(context.Background(), "", "")
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *system, httpResp, nil, 7)

	testutil.MockOnListManagerVirtualMedia(ctx, m, testutil.ManagerID,
		testutil.GetMediaCollection([]string{"Cd", "DVD", "Floppy"}), httpResp, nil, -1)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd", testMediaCD, httpResp, nil)
	testutil.MockOnEjectVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testutil.GetVirtualMedia([]string{"Cd"}), httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "DVD",
		testutil.GetVirtualMedia([]string{"DVD"}), httpResp, nil)
	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Floppy",
		testutil.GetVirtualMedia([]string{"Floppy"}), httpResp, nil)

	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testMediaCD, httpResp, nil)
	testutil.MockOnInsertVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		redfishClient.RedfishError{}, httpResp, redfishClient.GenericOpenAPIError{})

	testutil.MockOnGetManagerVirtualMedia(ctx, m, testutil.ManagerID, "Cd",
		testMediaCD, httpResp, nil)

	testutil.MockOnSetSystem(ctx, m, client.nodeID, redfishClient.ComputerSystem{},
		httpResp, nil)

	system.SetPowerState(redfishClient.POWERSTATE_ON)

	testutil.MockOnGetSystem(ctx, m, client.nodeID, *system, httpResp, nil, 1)

	resetReq := redfishClient.NewResetRequestBodyWithDefaults()
	resetReq.SetResetType(redfishClient.RESETTYPE_ON)
	testutil.MockOnResetSystem(ctx, m, client.nodeID, resetReq,
		redfishClient.RedfishError{}, httpResp, nil)
	system.SetPowerState(redfishClient.POWERSTATE_ON)
	testutil.MockOnGetSystem(ctx, m, client.nodeID, *system, httpResp, nil, 2)

	err = client.RemoteDirect(ctx, "http://some-url")
	assert.NoError(t, err)
}

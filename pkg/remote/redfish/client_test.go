package redfish

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"
)

const (
	ephemeralNodeID = "ephemeral-node-id"
	isoPath         = "https://localhost:8080/debian.iso"
	redfishURL      = "https://localhost:1234"
)

func getTestSystem() redfishClient.ComputerSystem {
	return redfishClient.ComputerSystem{
		Id:   "serverid-00",
		Name: "server-100",
		UUID: "58893887-8974-2487-2389-841168418919",
		Status: redfishClient.Status{
			State:  "Enabled",
			Health: "OK",
		},
		Links: redfishClient.SystemLinks{
			ManagedBy: []redfishClient.IdRef{
				{OdataId: "/redfish/v1/Managers/manager-1"},
			},
		},
		Boot: redfishClient.Boot{
			BootSourceOverrideTarget:  redfishClient.BOOTSOURCE_CD,
			BootSourceOverrideEnabled: redfishClient.BOOTSOURCEOVERRIDEENABLED_CONTINUOUS,
			BootSourceOverrideTargetRedfishAllowableValues: []redfishClient.BootSource{
				redfishClient.BOOTSOURCE_CD,
				redfishClient.BOOTSOURCE_FLOPPY,
				redfishClient.BOOTSOURCE_HDD,
				redfishClient.BOOTSOURCE_PXE,
			},
		},
	}
}

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
		nil, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx, "CD")
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSetEphemeralBootSourceByTypeSetSystemError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	m.On("GetSystem", ctx, client.ephemeralNodeID).Return(getTestSystem(),
		&http.Response{StatusCode: 200}, nil)
	m.On("SetSystem", ctx, client.ephemeralNodeID, mock.Anything).Times(1).Return(
		redfishClient.ComputerSystem{}, &http.Response{StatusCode: 401}, redfishClient.GenericOpenAPIError{})

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx, "CD")
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSetEphemeralBootSourceByTypeBootSourceUnavailable(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient("invalid-server", isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	invalidSystem := getTestSystem()
	invalidSystem.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	}

	m.On("GetSystem", ctx, client.ephemeralNodeID).Return(invalidSystem, nil, nil)

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetEphemeralBootSourceByType(ctx, "Cd")
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

	err = client.SetVirtualMedia(ctx, "CD", client.isoPath)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestSetVirtualMediaInsertVirtualMediaError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx, client, err := NewClient(systemID, isoPath, redfishURL, false, false, "", "")
	assert.NoError(t, err)

	httpResp := &http.Response{StatusCode: 500}
	m.On("GetSystem", context.Background(), systemID).Times(1).Return(getTestSystem(), nil, nil)

	realErr := redfishClient.GenericOpenAPIError{}
	m.On("InsertVirtualMedia", context.Background(), "manager-1", "Cd", mock.Anything).Return(
		redfishClient.RedfishError{}, httpResp, realErr)

	// Replace normal API client with mocked API client
	client.redfishAPI = m

	err = client.SetVirtualMedia(ctx, "Cd", client.isoPath)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

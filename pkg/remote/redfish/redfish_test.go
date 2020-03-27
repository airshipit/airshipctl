package redfish_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	redfishAPI "opendev.org/airship/go-redfish/api"
	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"

	. "opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	computerSystemID = "server-100"
	defaultURL       = "https://localhost:1234"
)

func TestRedfishRemoteDirectNormal(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	httpResp := &http.Response{StatusCode: 200}

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	m.On("GetSystem", ctx, systemID).Times(1).
		Return(getTestSystem(), httpResp, nil)
	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, systemID).Times(1).
		Return(getTestSystem(), httpResp, nil)
	systemReq := redfishClient.ComputerSystem{
		Boot: redfishClient.Boot{
			BootSourceOverrideTarget: redfishClient.BOOTSOURCE_CD,
		},
	}
	m.On("SetSystem", ctx, systemID, systemReq).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, nil)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, systemID).Times(1).
		Return(redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_OFF}, httpResp, nil)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, nil)

	m.On("GetSystem", ctx, systemID).Times(1).
		Return(redfishClient.ComputerSystem{PowerState: redfishClient.POWERSTATE_ON}, httpResp, nil)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	assert.NoError(t, err)
}

func TestRedfishRemoteDirectInvalidSystemId(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := "invalid-server"
	localRDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	localRDCfg.EphemeralNodeID = systemID

	realErr := fmt.Errorf("%s system do not exist", systemID)
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectGetSystemNetworkError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := computerSystemID
	realErr := fmt.Errorf("server request timeout")
	httpResp := &http.Response{StatusCode: 408}
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectInvalidIsoPath(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := computerSystemID
	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)
	localRDCfg := rDCfg
	localRDCfg.IsoPath = "bogus/path/to.iso"

	realErr := redfishClient.GenericOpenAPIError{}
	httpResp := &http.Response{StatusCode: 500}
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(getTestSystem(), nil, nil)

	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectCdDvdNotAvailableInBootSources(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := computerSystemID
	invalidSystem := getTestSystem()
	invalidSystem.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	}

	m.On("GetSystem", ctx, systemID).
		Return(invalidSystem, nil, nil)

	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSetSystemBootSourceFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := computerSystemID
	httpSuccResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, systemID).
		Return(getTestSystem(), httpSuccResp, nil)

	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, httpSuccResp, nil)

	m.On("SetSystem", ctx, systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, &http.Response{StatusCode: 401},
			redfishClient.GenericOpenAPIError{})

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSystemRebootFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.WithValue(
		context.Background(),
		redfishClient.ContextBasicAuth,
		redfishClient.BasicAuth{UserName: "username", Password: "password"},
	)

	systemID := computerSystemID
	httpSuccResp := &http.Response{StatusCode: 200}
	m.On("GetSystem", ctx, systemID).
		Return(getTestSystem(), httpSuccResp, nil)

	m.On("InsertVirtualMedia", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(redfishClient.RedfishError{}, httpSuccResp, nil)

	m.On("SetSystem", ctx, systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpSuccResp, nil)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 401},
			redfishClient.GenericOpenAPIError{})

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func getTestSystem() redfishClient.ComputerSystem {
	return redfishClient.ComputerSystem{
		Id:   computerSystemID,
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

func TestNewRedfishRemoteDirectClient(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	_, err := NewRedfishRemoteDirectClient(
		defaultURL,
		computerSystemID,
		"username",
		"password",
		"/tmp/test.iso",
		true,
		false,
	)
	assert.NoError(t, err)

	// Test with empty remote URL
	_, err = NewRedfishRemoteDirectClient(
		"",
		computerSystemID,
		"username",
		"password",
		"/tmp/test.iso",
		false,
		false,
	)
	expectedError := "missing configuration: redfish remote url empty"
	assert.EqualError(t, err, expectedError)

	// Test with empty ephemeral NodeID
	_, err = NewRedfishRemoteDirectClient(
		defaultURL,
		"",
		"username",
		"password",
		"/tmp/test.iso",
		false,
		false,
	)
	expectedError = "missing configuration: redfish ephemeral node id empty"
	assert.EqualError(t, err, expectedError)

	// Test with empty Iso Path
	_, err = NewRedfishRemoteDirectClient(
		defaultURL,
		computerSystemID,
		"username",
		"password",
		"",
		false,
		false,
	)
	expectedError = "missing configuration: redfish ephemeral node iso Path empty"
	assert.EqualError(t, err, expectedError)
}

func getDefaultRedfishRemoteDirectObj(t *testing.T, api redfishAPI.RedfishAPI) RemoteDirect {
	t.Helper()

	rDCfg, err := NewRedfishRemoteDirectClient(
		defaultURL,
		computerSystemID,
		"username",
		"password",
		"/tmp/test.iso",
		false,
		false,
	)

	require.NoError(t, err)

	rDCfg.RedfishAPI = api

	return rDCfg
}

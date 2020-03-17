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
	m.On("GetSystem", context.Background(), systemID).
		Return(getTestSystem(), nil, nil)
	m.On("InsertVirtualMedia", context.Background(), "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	systemReq := redfishClient.ComputerSystem{
		Boot: redfishClient.Boot{
			BootSourceOverrideTarget: redfishClient.BOOTSOURCE_CD,
		},
	}
	m.On("SetSystem", context.Background(), systemID, systemReq).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, nil)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", context.Background(), systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", context.Background(), systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	assert.NoError(t, err)
}

func TestRedfishRemoteDirectInvalidSystemId(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := "invalid-server"
	localRDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	localRDCfg.EphemeralNodeID = systemID

	realErr := fmt.Errorf("%s system do not exist", systemID)
	m.On("GetSystem", context.Background(), systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectGetSystemNetworkError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	realErr := fmt.Errorf("server request timeout")
	httpResp := &http.Response{
		StatusCode: 408,
	}
	m.On("GetSystem", context.Background(), systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectInvalidIsoPath(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)
	localRDCfg := rDCfg
	localRDCfg.IsoPath = "bogus/path/to.iso"

	errStr := "invalid remote boot path"
	realErr := fmt.Errorf(errStr)
	httpResp := &http.Response{
		StatusCode: 500,
	}
	m.On("GetSystem", context.Background(), systemID).
		Times(1).
		Return(getTestSystem(), nil, nil)

	m.On("InsertVirtualMedia", context.Background(), "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectCdDvdNotAvailableInBootSources(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	invalidSystem := getTestSystem()
	invalidSystem.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	}

	m.On("GetSystem", context.Background(), systemID).
		Return(invalidSystem, nil, nil)

	m.On("InsertVirtualMedia", context.Background(), "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSetSystemBootSourceFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	m.On("GetSystem", context.Background(), systemID).
		Return(getTestSystem(), nil, nil)

	m.On("InsertVirtualMedia", context.Background(), "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	realErr := fmt.Errorf("unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	m.On("SetSystem", context.Background(), systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSystemRebootFailed(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID

	m.On("GetSystem", context.Background(), systemID).
		Return(getTestSystem(), nil, nil)

	m.On("InsertVirtualMedia", context.Background(), mock.Anything, mock.Anything, mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	m.On("SetSystem", context.Background(), systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, nil)

	realErr := fmt.Errorf("unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", context.Background(), systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*ClientError)
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
		context.Background(),
		defaultURL,
		computerSystemID,
		"/tmp/test.iso",
	)
	assert.NoError(t, err)

	// Test with empty remote URL
	_, err = NewRedfishRemoteDirectClient(
		context.Background(),
		"",
		computerSystemID,
		"/tmp/test.iso",
	)
	expectedError := "missing configuration: redfish remote url empty"
	assert.EqualError(t, err, expectedError)

	// Test with empty ephemeral NodeID
	_, err = NewRedfishRemoteDirectClient(
		context.Background(),
		defaultURL,
		"",
		"/tmp/test.iso",
	)
	expectedError = "missing configuration: redfish ephemeral node id empty"
	assert.EqualError(t, err, expectedError)

	// Test with empty Iso Path
	_, err = NewRedfishRemoteDirectClient(
		context.Background(),
		defaultURL,
		computerSystemID,
		"",
	)
	expectedError = "missing configuration: redfish ephemeral node iso Path empty"
	assert.EqualError(t, err, expectedError)
}

func getDefaultRedfishRemoteDirectObj(t *testing.T, api redfishAPI.RedfishAPI) RemoteDirect {
	t.Helper()

	rDCfg, err := NewRedfishRemoteDirectClient(
		context.Background(),
		defaultURL,
		computerSystemID,
		"/tmp/test.iso",
	)

	require.NoError(t, err)

	rDCfg.RedfishAPI = api

	return rDCfg
}

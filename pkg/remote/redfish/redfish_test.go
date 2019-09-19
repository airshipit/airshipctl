package redfish_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	redfishAPI "github.com/Nordix/go-redfish/api"
	redfishMocks "github.com/Nordix/go-redfish/api/mocks"
	redfishClient "github.com/Nordix/go-redfish/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	. "opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	computerSystemID = "server-100"
)

var (
	ctx        = context.Background()
	defaultURL = "https://localhost:1234"

	SampleSystem = redfishClient.ComputerSystem{
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
)

func getDefaultRedfishRemoteDirectObj(t *testing.T, api redfishAPI.RedfishAPI) RedfishRemoteDirect {

	t.Helper()

	rDCfg, err := NewRedfishRemoteDirectClient(
		ctx,
		defaultURL,
		computerSystemID,
		"/tmp/test.iso",
	)

	assert.NoError(t, err)

	rDCfg.Api = api

	return rDCfg
}

func TestRedfishRemoteDirectNormal(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	m.On("GetSystem", ctx, systemID).
		Return(SampleSystem, nil, nil)
	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	systemReq := redfishClient.ComputerSystem{
		Boot: redfishClient.Boot{
			BootSourceOverrideTarget: redfishClient.BOOTSOURCE_CD,
		},
	}
	m.On("SetSystem", ctx, systemID, systemReq).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, nil)

	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetReq).
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

	localRDCfg.EphemeralNodeId = systemID

	realErr := fmt.Errorf("%s system do not exist", systemID)
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectGetSystemNetworkError(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	realErr := fmt.Errorf("Server request timeout")
	httpResp := &http.Response{
		StatusCode: 408,
	}
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectInvalidIsoPath(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)
	localRDCfg := rDCfg
	localRDCfg.IsoPath = "bogus/path/to.iso"

	errStr := "Invalid remote boot path"
	realErr := fmt.Errorf(errStr)
	httpResp := &http.Response{
		StatusCode: 500,
	}
	m.On("GetSystem", ctx, systemID).
		Times(1).
		Return(SampleSystem, nil, nil)

	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := localRDCfg.DoRemoteDirect()

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectCdDvdNotAvailableInBootSources(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	invalidSystem := SampleSystem
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

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSetSystemBootSourceFailed(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID
	m.On("GetSystem", ctx, systemID).
		Return(SampleSystem, nil, nil)

	m.On("InsertVirtualMedia", ctx, "manager-1", "Cd", mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	realErr := fmt.Errorf("Unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	m.On("SetSystem", ctx, systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishRemoteDirectSystemRebootFailed(t *testing.T) {

	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	systemID := computerSystemID

	m.On("GetSystem", ctx, systemID).
		Return(SampleSystem, nil, nil)

	m.On("InsertVirtualMedia", ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(redfishClient.RedfishError{}, nil, nil)

	m.On("SetSystem", ctx, systemID, mock.Anything).
		Times(1).
		Return(redfishClient.ComputerSystem{}, nil, nil)

	realErr := fmt.Errorf("Unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	rDCfg := getDefaultRedfishRemoteDirectObj(t, m)

	err := rDCfg.DoRemoteDirect()

	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

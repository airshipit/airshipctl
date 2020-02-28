package redfish

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"
)

const (
	systemID = "123"
)

func TestRedfishErrorNoError(t *testing.T) {
	err := ScreenRedfishError(nil, nil)
	assert.NoError(t, err)
}

func TestRedfishErrorNonNilErrorWithoutHttpResp(t *testing.T) {
	realErr := fmt.Errorf("sample error")
	err := ScreenRedfishError(nil, realErr)
	assert.Error(t, err)
	_, ok := err.(*RedfishClientError)
	assert.True(t, ok)
}

func TestRedfishErrorNonNilErrorWithHttpRespError(t *testing.T) {
	realErr := fmt.Errorf("sample error")

	httpResp := &http.Response{StatusCode: 408}
	err := ScreenRedfishError(httpResp, realErr)
	assert.Equal(t, err, NewRedfishClientErrorf(realErr.Error()))

	httpResp.StatusCode = 400
	err = ScreenRedfishError(httpResp, realErr)
	assert.Equal(t, err, NewRedfishClientErrorf(realErr.Error()))

	httpResp.StatusCode = 199
	err = ScreenRedfishError(httpResp, realErr)
	assert.Equal(t, err, NewRedfishClientErrorf(realErr.Error()))
}

func TestRedfishErrorNonNilErrorWithHttpRespOK(t *testing.T) {
	realErr := fmt.Errorf("sample error")

	httpResp := &http.Response{StatusCode: 204}
	err := ScreenRedfishError(httpResp, realErr)
	assert.NoError(t, err)

	httpResp.StatusCode = 200
	err = ScreenRedfishError(httpResp, realErr)
	assert.NoError(t, err)

	httpResp.StatusCode = 399
	err = ScreenRedfishError(httpResp, realErr)
	assert.NoError(t, err)
}

func TestRedfishUtilGetResIDFromURL(t *testing.T) {
	// simple case
	url := "api/user/123"
	id := GetResourceIDFromURL(url)
	assert.Equal(t, id, "123")

	// FQDN
	url = "http://www.abc.com/api/user/123"
	id = GetResourceIDFromURL(url)
	assert.Equal(t, id, "123")

	//Trailing slash
	url = "api/user/123/"
	id = GetResourceIDFromURL(url)
	assert.Equal(t, id, "123")
}

func TestRedfishUtilIsIdInList(t *testing.T) {
	idList := []redfishClient.IdRef{
		{OdataId: "/path/to/id/1"},
		{OdataId: "/path/to/id/2"},
		{OdataId: "/path/to/id/3"},
		{OdataId: "/path/to/id/4"},
	}
	var emptyList []redfishClient.IdRef

	res := IsIDInList(idList, "1")
	assert.True(t, res)

	res = IsIDInList(idList, "100")
	assert.False(t, res)

	res = IsIDInList(emptyList, "1")
	assert.False(t, res)
}

func TestRedfishUtilRebootSystemOK(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	err := RebootSystem(ctx, m, systemID)
	assert.NoError(t, err)
}

func TestRedfishUtilRebootSystemForceOffError2(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	realErr := fmt.Errorf("unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := RebootSystem(ctx, m, systemID)
	assert.Error(t, err)
}

func TestRedfishUtilRebootSystemForceOffError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	realErr := fmt.Errorf("unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := RebootSystem(ctx, m, systemID)
	assert.Error(t, err)
}

func TestRedfishUtilRebootSystemTurningOnError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, nil, nil)

	realErr := fmt.Errorf("unauthorized")
	httpResp := &http.Response{
		StatusCode: 401,
	}
	resetOnReq := redfishClient.ResetRequestBody{}
	resetOnReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetOnReq).
		Times(1).
		Return(redfishClient.RedfishError{}, httpResp, realErr)

	err := RebootSystem(ctx, m, systemID)
	assert.Error(t, err)
}

package redfish

import (
	"context"
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
	err := ScreenRedfishError(&http.Response{StatusCode: 200}, nil)
	assert.NoError(t, err)
}

func TestRedfishErrorNonNilErrorWithoutHttpResp(t *testing.T) {
	err := ScreenRedfishError(nil, redfishClient.GenericOpenAPIError{})

	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishErrorNonNilErrorWithHttpRespError(t *testing.T) {
	respErr := redfishClient.GenericOpenAPIError{}

	err := ScreenRedfishError(&http.Response{StatusCode: 408}, respErr)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)

	err = ScreenRedfishError(&http.Response{StatusCode: 500}, respErr)
	_, ok = err.(ErrRedfishClient)
	assert.True(t, ok)

	err = ScreenRedfishError(&http.Response{StatusCode: 199}, respErr)
	_, ok = err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishErrorNonNilErrorWithHttpRespOK(t *testing.T) {
	respErr := redfishClient.GenericOpenAPIError{}

	// NOTE: Redfish client only uses HTTP 200 & HTTP 204 for success.
	err := ScreenRedfishError(&http.Response{StatusCode: 204}, respErr)
	assert.NoError(t, err)

	err = ScreenRedfishError(&http.Response{StatusCode: 200}, respErr)
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
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	resetReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	err := RebootSystem(ctx, m, systemID)
	assert.NoError(t, err)
}

func TestRedfishUtilRebootSystemForceOffError2(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF

	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 401},
			redfishClient.GenericOpenAPIError{})

	err := RebootSystem(ctx, m, systemID)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishUtilRebootSystemForceOffError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 401},
			redfishClient.GenericOpenAPIError{})

	err := RebootSystem(ctx, m, systemID)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishUtilRebootSystemTurningOnError(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	resetReq := redfishClient.ResetRequestBody{}
	resetReq.ResetType = redfishClient.RESETTYPE_FORCE_OFF
	m.On("ResetSystem", ctx, systemID, resetReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 200}, nil)

	resetOnReq := redfishClient.ResetRequestBody{}
	resetOnReq.ResetType = redfishClient.RESETTYPE_ON
	m.On("ResetSystem", ctx, systemID, resetOnReq).
		Times(1).
		Return(redfishClient.RedfishError{}, &http.Response{StatusCode: 401},
			redfishClient.GenericOpenAPIError{})

	err := RebootSystem(ctx, m, systemID)
	_, ok := err.(ErrRedfishClient)
	assert.True(t, ok)
}

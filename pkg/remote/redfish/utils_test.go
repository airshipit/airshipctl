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

package redfish_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"

	"opendev.org/airship/airshipctl/pkg/remote/redfish"
	testutil "opendev.org/airship/airshipctl/testutil/redfishutils/helpers"
)

const (
	redfishHTTPErrDMTF = `
{
  "error": {
    "message": "A general error has occurred. See Resolution for information on how to resolve the error.",
    "@Message.ExtendedInfo": [
      {
        "Message": "Extended error message.",
        "Resolution": "Resolution message."
      },
      {
        "Message": "Extended message 2.",
        "Resolution": "Resolution message 2."
      }
    ]
  }
}`
	redfishHTTPErrOther = `
{
  "error": {
    "message": "A general error has occurred. See Resolution for information on how to resolve the error.",
    "@Message.ExtendedInfo": {
      "Message": "Extended error message.",
      "Resolution": "Resolution message."
    }
  }
}`
	redfishHTTPErrMalformatted = `
{
  "error": {
    "message": "A general error has occurred. See Resolution for information on how to resolve the error.",
    "@Message.ExtendedInfo": {}
  }
}`
	redfishHTTPErrEmptyList = `
{
  "error": {
    "message": "A general error has occurred. See Resolution for information on how to resolve the error.",
    "@Message.ExtendedInfo": []
  }
}`
)

func TestDecodeRawErrorEmptyInput(t *testing.T) {
	_, err := redfish.DecodeRawError([]byte("{}"))
	assert.Error(t, err)
}

func TestDecodeRawErrorEmptyList(t *testing.T) {
	_, err := redfish.DecodeRawError([]byte(redfishHTTPErrEmptyList))
	assert.Error(t, err)
}

func TestDecodeRawErrorEmptyMalformatted(t *testing.T) {
	_, err := redfish.DecodeRawError([]byte(redfishHTTPErrMalformatted))
	assert.Error(t, err)
}

func TestDecodeRawErrorDMTF(t *testing.T) {
	message, err := redfish.DecodeRawError([]byte(redfishHTTPErrDMTF))
	assert.NoError(t, err)
	assert.Equal(t, "Extended message 2. Resolution message 2.\nExtended error message. Resolution message.\n",
		message)
}

func TestDecodeRawErrorOther(t *testing.T) {
	message, err := redfish.DecodeRawError([]byte(redfishHTTPErrOther))
	assert.NoError(t, err)
	assert.Equal(t, "Extended error message. Resolution message.",
		message)
}

func TestRedfishErrorNoError(t *testing.T) {
	err := redfish.ScreenRedfishError(&http.Response{StatusCode: 200}, nil)
	assert.NoError(t, err)
}

func TestRedfishErrorNonNilErrorWithoutHttpResp(t *testing.T) {
	err := redfish.ScreenRedfishError(nil, redfishClient.GenericOpenAPIError{})

	_, ok := err.(redfish.ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishErrorNonNilErrorWithHttpRespError(t *testing.T) {
	respErr := redfishClient.GenericOpenAPIError{}

	err := redfish.ScreenRedfishError(&http.Response{StatusCode: 408}, respErr)
	_, ok := err.(redfish.ErrRedfishClient)
	assert.True(t, ok)

	err = redfish.ScreenRedfishError(&http.Response{StatusCode: 500}, respErr)
	_, ok = err.(redfish.ErrRedfishClient)
	assert.True(t, ok)

	err = redfish.ScreenRedfishError(&http.Response{StatusCode: 199}, respErr)
	_, ok = err.(redfish.ErrRedfishClient)
	assert.True(t, ok)
}

func TestRedfishErrorNonNilErrorWithHttpRespOK(t *testing.T) {
	respErr := redfishClient.GenericOpenAPIError{}

	// NOTE: Redfish client only uses HTTP 200 & HTTP 204 for success.
	err := redfish.ScreenRedfishError(&http.Response{StatusCode: 204}, respErr)
	assert.NoError(t, err)

	err = redfish.ScreenRedfishError(&http.Response{StatusCode: 200}, respErr)
	assert.NoError(t, err)
}

func TestRedfishUtilGetResIDFromURL(t *testing.T) {
	// simple case
	url := "api/user/123"
	id := redfish.GetResourceIDFromURL(url)
	assert.Equal(t, id, "123")

	// FQDN
	url = "http://www.abc.com/api/user/123"
	id = redfish.GetResourceIDFromURL(url)
	assert.Equal(t, id, "123")

	//Trailing slash
	url = "api/user/123/"
	id = redfish.GetResourceIDFromURL(url)
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

	res := redfish.IsIDInList(idList, "1")
	assert.True(t, res)

	res = redfish.IsIDInList(idList, "100")
	assert.False(t, res)

	res = redfish.IsIDInList(emptyList, "1")
	assert.False(t, res)
}

func TestGetVirtualMediaID(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	httpResp := &http.Response{StatusCode: 200}

	m.On("GetSystem", ctx, mock.Anything).
		Return(testutil.GetTestSystem(), &http.Response{StatusCode: 200}, nil)

	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Floppy", "Cd"}), httpResp, nil)

	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Floppy").Times(1).
		Return(testutil.GetVirtualMedia([]string{"Floppy", "USBStick"}), httpResp, nil)

	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Cd").Times(1).
		Return(testutil.GetVirtualMedia([]string{"CD"}), httpResp, nil)

	mediaID, mediaType, err := redfish.GetVirtualMediaID(ctx, m, testutil.ManagerID)
	assert.Equal(t, mediaID, "Cd")
	assert.Equal(t, mediaType, "CD")
	assert.NoError(t, err)
}

func TestGetVirtualMediaIDNoMedia(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	httpResp := &http.Response{StatusCode: 200}

	// Remove available media types from test system
	system := testutil.GetTestSystem()
	system.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{}
	m.On("GetSystem", ctx, mock.Anything).
		Return(system, &http.Response{StatusCode: 200}, nil)

	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(redfishClient.Collection{}, httpResp, nil)

	mediaID, mediaType, err := redfish.GetVirtualMediaID(ctx, m, testutil.ManagerID)
	assert.Empty(t, mediaID)
	assert.Empty(t, mediaType)
	assert.Error(t, err)
}

func TestGetVirtualMediaIDUnacceptableMediaTypes(t *testing.T) {
	m := &redfishMocks.RedfishAPI{}
	defer m.AssertExpectations(t)

	ctx := context.Background()
	httpResp := &http.Response{StatusCode: 200}

	system := testutil.GetTestSystem()
	system.Boot.BootSourceOverrideTargetRedfishAllowableValues = []redfishClient.BootSource{
		redfishClient.BOOTSOURCE_PXE,
	}
	m.On("GetSystem", ctx, mock.Anything).
		Return(system, &http.Response{StatusCode: 200}, nil)

	m.On("ListManagerVirtualMedia", ctx, testutil.ManagerID).Times(1).
		Return(testutil.GetMediaCollection([]string{"Floppy"}), httpResp, nil)

	m.On("GetManagerVirtualMedia", ctx, testutil.ManagerID, "Floppy").Times(1).
		Return(testutil.GetVirtualMedia([]string{"Floppy", "USBStick"}), httpResp, nil)

	mediaID, mediaType, err := redfish.GetVirtualMediaID(ctx, m, testutil.ManagerID)
	assert.Empty(t, mediaID)
	assert.Empty(t, mediaType)
	assert.Error(t, err)
}

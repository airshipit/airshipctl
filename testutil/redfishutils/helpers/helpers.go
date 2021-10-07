// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package redfishutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/mock"

	redfishMocks "opendev.org/airship/go-redfish/api/mocks"
	redfishClient "opendev.org/airship/go-redfish/client"
)

const (
	// ManagerID is the Redfish manager ID used by helper functions and should be used in mock calls.
	ManagerID = "manager1"
)

// GetMediaCollection builds a collection of media IDs returned by the "ListManagerVirtualMedia" function.
func GetMediaCollection(refs []string) redfishClient.Collection {
	uri := "/redfish/v1/Managers/7832-09/VirtualMedia"
	ids := []redfishClient.IdRef{}

	for _, r := range refs {
		id := redfishClient.IdRef{}
		id.SetOdataId(fmt.Sprintf("%s/%s", uri, r))
		ids = append(ids, id)
	}

	c := redfishClient.Collection{Members: ids}

	return c
}

// GetVirtualMedia builds an array of virtual media resources returned by the "GetManagerVirtualMedia" function.
func GetVirtualMedia(types []string) redfishClient.VirtualMedia {
	vMedia := redfishClient.VirtualMedia{}

	mediaTypes := []string{}
	for _, t := range types {
		mediaTypes = append(mediaTypes, t)
	}

	inserted := false

	vMedia.SetMediaTypes(mediaTypes)
	vMedia.SetInserted(inserted)

	return vMedia
}

// GetTestSystem builds a test computer system.
func GetTestSystem() redfishClient.ComputerSystem {
	computerSystem := redfishClient.NewComputerSystem()
	computerSystem.SetId("serverid-00")
	computerSystem.SetName("server-100")
	computerSystem.SetUUID("58893887-8974-2487-2389-841168418919")

	status := redfishClient.NewStatusWithDefaults()
	status.SetState(redfishClient.STATE_ENABLED)
	status.SetHealth(redfishClient.HEALTH_OK)
	computerSystem.SetStatus(*status)

	sysLinks := redfishClient.NewSystemLinksWithDefaults()
	sysLinks.SetManagedBy([]redfishClient.IdRef{
		{
			OdataId: redfishClient.PtrString(fmt.Sprintf("/redfish/v1/Managers/%s", ManagerID)),
		},
	})
	computerSystem.SetLinks(*sysLinks)
	boot := redfishClient.NewBoot()
	boot.SetBootSourceOverrideTarget(redfishClient.BOOTSOURCE_CD)
	boot.SetBootSourceOverrideEnabled(redfishClient.BOOTSOURCEOVERRIDEENABLED_CONTINUOUS)
	boot.SetBootSourceOverrideTargetRedfishAllowableValues([]redfishClient.BootSource{
		redfishClient.BOOTSOURCE_CD,
		redfishClient.BOOTSOURCE_FLOPPY,
		redfishClient.BOOTSOURCE_HDD,
		redfishClient.BOOTSOURCE_PXE,
	})
	computerSystem.SetBoot(*boot)
	return *computerSystem
}

// MockOnGetSystem creates mock On calls for GetSystem and GetSystemExecute
func MockOnGetSystem(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	systemID string, computerSystem redfishClient.ComputerSystem,
	httpResponse *http.Response, err error, times int) {
	testSystemRequest := redfishClient.ApiGetSystemRequest{}
	call := mockAPI.On("GetSystem", ctx, systemID).Return(testSystemRequest)
	if times > 0 {
		call.Times(times)
	}
	call = mockAPI.On("GetSystemExecute", testSystemRequest).Return(computerSystem, httpResponse, err)
	if times > 0 {
		call.Times(times)
	}
}

// MockOnResetSystem creates mock On calls for ResetSystem and ResetSystemExecute
func MockOnResetSystem(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	systemID string, requestBody *redfishClient.ResetRequestBody, redfishErr redfishClient.RedfishError,
	httpResponse *http.Response, err error) {
	request := redfishClient.ApiResetSystemRequest{}.ResetRequestBody(*requestBody)
	mockAPI.On("ResetSystem", ctx, systemID).Return(request).Times(1)
	mockAPI.On("ResetSystemExecute", mock.Anything).Return(redfishErr, httpResponse, err).Times(1)
}

// MockOnSetSystem creates mock On calls for SetSystem and SetSystemExecute
func MockOnSetSystem(ctx context.Context, mockAPI *redfishMocks.RedfishAPI, systemID string,
	computerSystem redfishClient.ComputerSystem, httpResponse *http.Response, err error) {
	request := redfishClient.ApiSetSystemRequest{}.ComputerSystem(computerSystem)
	mockAPI.On("SetSystem", ctx, systemID).Return(request).Times(1)
	mockAPI.On("SetSystemExecute", mock.Anything).Return(computerSystem, httpResponse, err).Times(1)
}

// MockOnGetManagerVirtualMedia creates mock On calls for GetManagerVirtualMedia and GetManagerVirtualMediaExecute
func MockOnGetManagerVirtualMedia(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	managerID string, virtualMediaID string, virtualMedia redfishClient.VirtualMedia,
	httpResponse *http.Response, err error) {
	mediaRequest := redfishClient.ApiGetManagerVirtualMediaRequest{}
	mockAPI.On("GetManagerVirtualMedia", ctx, managerID, virtualMediaID).Return(mediaRequest).Times(1)
	mockAPI.On("GetManagerVirtualMediaExecute", mock.Anything).Return(virtualMedia, httpResponse, err).Times(1)
}

// MockOnListManagerVirtualMedia creates mock On calls for ListManagerVirtualMedia and ListtManagerVirtualMediaExecute
func MockOnListManagerVirtualMedia(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	managerID string, collection redfishClient.Collection, httpResponse *http.Response, err error, times int) {
	mediaRequest := redfishClient.ApiListManagerVirtualMediaRequest{}
	called := mockAPI.On("ListManagerVirtualMedia", ctx, managerID).Return(mediaRequest)
	if times > 0 {
		called.Times(times)
	}
	called = mockAPI.On("ListManagerVirtualMediaExecute", mock.Anything).Return(collection, httpResponse, err)
	if times > 0 {
		called.Times(1)
	}
}

// MockOnEjectVirtualMedia creates mock On calls for EjectVirtualMedia and EjectVirtualMediaExecute
func MockOnEjectVirtualMedia(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	managerID string, virtualMediaID string, redfishErr redfishClient.RedfishError,
	httpResponse *http.Response, err error) {
	mediaRequest := redfishClient.ApiEjectVirtualMediaRequest{}
	mockAPI.On("EjectVirtualMedia", ctx, managerID, virtualMediaID).Return(mediaRequest).Times(1)
	mockAPI.On("EjectVirtualMediaExecute", mock.Anything).Return(redfishErr, httpResponse, err).Times(1)
}

// MockOnInsertVirtualMedia creates mock On calls for InsertVirtualMedia and InsertVirtualMediaExecute
func MockOnInsertVirtualMedia(ctx context.Context, mockAPI *redfishMocks.RedfishAPI,
	managerID string, virtualMediaID string, redfishErr redfishClient.RedfishError,
	httpResponse *http.Response, err error) {
	mediaRequest := redfishClient.ApiInsertVirtualMediaRequest{}
	mockAPI.On("InsertVirtualMedia", ctx, managerID, virtualMediaID).Return(mediaRequest).Times(1)
	mockAPI.On("InsertVirtualMediaExecute", mock.Anything).Return(redfishErr, httpResponse, err).Times(1)
}

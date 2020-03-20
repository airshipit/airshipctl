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
	"fmt"

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
		id.OdataId = fmt.Sprintf("%s/%s", uri, r)
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

	vMedia.MediaTypes = mediaTypes

	return vMedia
}

// GetTestSystem builds a test computer system.
func GetTestSystem() redfishClient.ComputerSystem {
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
				{OdataId: fmt.Sprintf("/redfish/v1/Managers/%s", ManagerID)},
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

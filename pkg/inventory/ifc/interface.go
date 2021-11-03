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

package ifc

import (
	"context"

	remoteifc "opendev.org/airship/airshipctl/pkg/remote/ifc"
)

// Inventory interface for airshipctl
type Inventory interface {
	BaremetalInventory() (BaremetalInventory, error)
}

// BaremetalInventory interface that allows working with baremetal hosts
type BaremetalInventory interface {
	Select(BaremetalHostSelector) ([]remoteifc.Client, error)
	SelectOne(BaremetalHostSelector) (remoteifc.Client, error)
	RunOperation(context.Context, BaremetalOperation, BaremetalHostSelector, BaremetalBatchRunOptions) error
}

// BaremetalOperation baremetal operation
type BaremetalOperation string

const (
	// BaremetalOperationReboot reboot
	BaremetalOperationReboot BaremetalOperation = "reboot"
	// BaremetalOperationPowerOff power off
	BaremetalOperationPowerOff BaremetalOperation = "power-off"
	// BaremetalOperationPowerOn power on
	BaremetalOperationPowerOn BaremetalOperation = "power-on"
	// BaremetalOperationEjectVirtualMedia eject virtual media
	BaremetalOperationEjectVirtualMedia BaremetalOperation = "eject-virtual-media"
	// BaremetalOperationListHosts list hosts
	BaremetalOperationListHosts BaremetalOperation = "list-hosts"
)

// BaremetalBatchRunOptions are options to be passed to RunOperation, this is to be
// exetended in the future, to support such things as concurency
type BaremetalBatchRunOptions struct {
}

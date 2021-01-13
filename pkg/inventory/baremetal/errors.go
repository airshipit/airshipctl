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

package baremetal

import (
	"fmt"

	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
)

// ErrRemoteDriverNotSupported is returned when remote driver is not supported for baremetal host
type ErrRemoteDriverNotSupported struct {
	BMHName      string
	BMHNamespace string
	RemoteType   string
}

func (e ErrRemoteDriverNotSupported) Error() string {
	return fmt.Sprintf("Baremetal host named '%s' in namespace '%s' remote driver '%s' not supported",
		e.BMHName, e.BMHNamespace, e.RemoteType)
}

// ErrNoBaremetalHostsFound is returned when no baremetal hosts matched the selector
type ErrNoBaremetalHostsFound struct {
	Selector ifc.BaremetalHostSelector
}

func (e ErrNoBaremetalHostsFound) Error() string {
	return fmt.Sprintf("No baremetal hosts matched selector %v", e.Selector)
}

// ErrBaremetalOperationNotSupported is returned when baremetal operation is not supported
type ErrBaremetalOperationNotSupported struct {
	Operation ifc.BaremetalOperation
}

func (e ErrBaremetalOperationNotSupported) Error() string {
	return fmt.Sprintf("Baremetal operation not supported: '%s'", e.Operation)
}

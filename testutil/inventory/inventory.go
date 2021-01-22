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

package inventory

import (
	"context"

	"github.com/stretchr/testify/mock"

	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
	remoteifc "opendev.org/airship/airshipctl/pkg/remote/ifc"
)

var _ ifc.Inventory = &MockInventory{}

// MockInventory mocks ifc.Inventory interface
type MockInventory struct {
	mock.Mock
}

// BaremetalInventory mock
func (i *MockInventory) BaremetalInventory() (ifc.BaremetalInventory, error) {
	args := i.Called()
	err := args.Error(1)
	bmhInv, ok := args.Get(0).(ifc.BaremetalInventory)
	if !ok {
		return nil, err
	}
	return bmhInv, err
}

var _ ifc.BaremetalInventory = &MockBMHInventory{}

// MockBMHInventory mocks ifc.BaremetalInventory
type MockBMHInventory struct {
	mock.Mock
}

// Select mock
func (i *MockBMHInventory) Select(ifc.BaremetalHostSelector) ([]remoteifc.Client, error) {
	args := i.Called()
	err := args.Error(1)
	hosts, ok := args.Get(0).([]remoteifc.Client)
	if !ok {
		return nil, err
	}
	return hosts, nil
}

// SelectOne mock
func (i *MockBMHInventory) SelectOne(ifc.BaremetalHostSelector) (remoteifc.Client, error) {
	args := i.Called()
	err := args.Error(1)
	host, ok := args.Get(0).(remoteifc.Client)
	if !ok {
		return nil, err
	}
	return host, nil
}

// RunOperation mock
func (i *MockBMHInventory) RunOperation(
	context.Context,
	ifc.BaremetalOperation,
	ifc.BaremetalHostSelector,
	ifc.BaremetalBatchRunOptions) error {
	return i.Called().Error(0)
}

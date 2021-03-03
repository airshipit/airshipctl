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

package inventory_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/inventory"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/remote/power"
	mockinventory "opendev.org/airship/airshipctl/testutil/inventory"
	"opendev.org/airship/airshipctl/testutil/redfishutils"
)

const testNode = "node-0"

func TestCommandOptions(t *testing.T) {
	t.Run("error BMHAction bmh inventory", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}
		expectedErr := fmt.Errorf("bmh inventory error")
		inv.On("BaremetalInventory").Once().Return(nil, expectedErr)

		co := inventory.NewOptions(inv)
		co.All = true
		actualErr := co.BMHAction(ifc.BaremetalOperationPowerOn)
		assert.Equal(t, expectedErr, actualErr)
	})

	t.Run("error BMHAction invalid empty options", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		co := inventory.NewOptions(inv)
		err := co.BMHAction(ifc.BaremetalOperationPowerOn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), (inventory.ErrInvalidOptions{}).Error())
	})

	t.Run("error BMHAction invalid both all and other selectors", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		co := inventory.NewOptions(inv)
		co.All = true
		co.Labels = "foo=bar"
		err := co.BMHAction(ifc.BaremetalOperationPowerOn)
		require.Error(t, err)
		assert.Contains(t, err.Error(), (inventory.ErrInvalidOptions{}).Error())
	})

	t.Run("success BMHAction", func(t *testing.T) {
		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("RunOperation").Once().Return(nil)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.All = true
		actualErr := co.BMHAction(ifc.BaremetalOperationPowerOn)
		assert.Equal(t, nil, actualErr)
	})

	t.Run("error PowerStatus SelectOne", func(t *testing.T) {
		expectedErr := fmt.Errorf("SelectOne inventory error")
		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("SelectOne").Once().Return(nil, expectedErr)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		buf := bytes.NewBuffer([]byte{})
		actualErr := co.PowerStatus(buf)
		assert.Equal(t, expectedErr, actualErr)
		assert.Len(t, buf.Bytes(), 0)
	})

	t.Run("error PowerStatus BMHInventory", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		expectedErr := fmt.Errorf("bmh inventory error")
		inv.On("BaremetalInventory").Once().Return(nil, expectedErr)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		buf := bytes.NewBuffer([]byte{})
		actualErr := co.PowerStatus(buf)
		assert.Equal(t, expectedErr, actualErr)
		assert.Len(t, buf.Bytes(), 0)
	})

	t.Run("error PowerStatus SystemPowerStatus", func(t *testing.T) {
		expectedErr := fmt.Errorf("SystemPowerStatus error")
		host := &redfishutils.MockClient{}
		host.On("SystemPowerStatus").Once().Return(power.StatusUnknown, expectedErr)

		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("SelectOne").Once().Return(host, nil)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		buf := bytes.NewBuffer([]byte{})
		actualErr := co.PowerStatus(buf)
		assert.Equal(t, expectedErr, actualErr)
		assert.Len(t, buf.Bytes(), 0)
	})

	t.Run("success PowerStatus", func(t *testing.T) {
		host := &redfishutils.MockClient{}
		nodeID := "node01"
		host.On("SystemPowerStatus").Once().Return(power.StatusPoweringOn, nil)
		host.On("NodeID").Once().Return(nodeID)

		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("SelectOne").Once().Return(host, nil)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		buf := bytes.NewBuffer([]byte{})
		actualErr := co.PowerStatus(buf)
		assert.Equal(t, nil, actualErr)
		assert.Contains(t, buf.String(), nodeID)
		assert.Contains(t, buf.String(), power.StatusPoweringOn.String())
	})

	t.Run("success RemoteDirect", func(t *testing.T) {
		host := &redfishutils.MockClient{}
		host.On("RemoteDirect").Once().Return(nil)

		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("SelectOne").Once().Return(host, nil)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		co.IsoURL = "http://some-url"
		actualErr := co.RemoteDirect()
		assert.Equal(t, nil, actualErr)
	})

	t.Run("error RemoteDirect no isoURL", func(t *testing.T) {
		host := &redfishutils.MockClient{}
		host.On("RemoteDirect").Once()

		bmhInv := &mockinventory.MockBMHInventory{}
		bmhInv.On("SelectOne").Once().Return(host, nil)

		inv := &mockinventory.MockInventory{}
		inv.On("BaremetalInventory").Once().Return(bmhInv, nil)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		actualErr := co.RemoteDirect()
		// Simply check if error is returned in isoURL is not specified
		assert.Error(t, actualErr)
	})

	t.Run("error RemoteDirect BMHInventory", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		expectedErr := fmt.Errorf("bmh inventory error")
		inv.On("BaremetalInventory").Once().Return(nil, expectedErr)

		co := inventory.NewOptions(inv)
		co.Name = testNode
		actualErr := co.RemoteDirect()
		assert.Equal(t, expectedErr, actualErr)
	})

	t.Run("error RemoteDirect invalid options", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		co := inventory.NewOptions(inv)
		err := co.RemoteDirect()
		require.Error(t, err)
		assert.Contains(t, err.Error(), (inventory.ErrInvalidOptions{}).Error())
	})

	t.Run("error RemoteDirect invalid options", func(t *testing.T) {
		inv := &mockinventory.MockInventory{}

		co := inventory.NewOptions(inv)
		buf := bytes.NewBuffer([]byte{})
		err := co.PowerStatus(buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), (inventory.ErrInvalidOptions{}).Error())
		assert.Len(t, buf.Bytes(), 0)
	})
}

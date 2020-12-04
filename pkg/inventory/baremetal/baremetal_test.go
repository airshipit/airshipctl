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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		name, remoteDriver, expectedErr string
		expectedHosts                   int

		selector ifc.BaremetalHostSelector
	}{
		{
			name:          "success return one host",
			remoteDriver:  "redfish-dell",
			expectedHosts: 1,
			selector:      (ifc.BaremetalHostSelector{}).ByName("master-0"),
		},
		{
			name:          "success return multiple host",
			remoteDriver:  "redfish",
			expectedHosts: 2,
			selector:      (ifc.BaremetalHostSelector{}).ByLabel("host-group=control-plane"),
		},
		{
			name:         "error remote driver not supported",
			remoteDriver: "should return error",
			expectedErr:  "not supported",
			selector:     (ifc.BaremetalHostSelector{}).ByLabel("host-group=control-plane"),
		},
		{
			name:         "error no credentials",
			remoteDriver: "redfish",
			expectedErr:  "no field named",
			selector:     (ifc.BaremetalHostSelector{}).ByName("no-creds"),
		},
		{
			name:          "error no hosts found",
			remoteDriver:  "redfish",
			expectedHosts: 0,
			selector:      (ifc.BaremetalHostSelector{}).ByName("no such host"),
		},
	}

	bundle := testBundle(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mgmCfg := config.ManagementConfiguration{Type: tt.remoteDriver}
			inventory := NewInventory(mgmCfg, bundle)
			hosts, err := inventory.Select(tt.selector)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Len(t, hosts, tt.expectedHosts)
			}
		})
	}
}

func TestSelectOne(t *testing.T) {
	tests := []struct {
		name, remoteDriver, expectedErr string

		selector ifc.BaremetalHostSelector
	}{
		{
			name:         "success return one host",
			remoteDriver: "redfish-dell",
			selector:     (ifc.BaremetalHostSelector{}).ByName("master-0"),
		},
		{
			name:         "error return multiple host",
			remoteDriver: "redfish",
			expectedErr:  "found more than one document",
			selector:     (ifc.BaremetalHostSelector{}).ByLabel("host-group=control-plane"),
		},
	}

	bundle := testBundle(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mgmCfg := config.ManagementConfiguration{Type: tt.remoteDriver}
			inventory := NewInventory(mgmCfg, bundle)
			host, err := inventory.SelectOne(tt.selector)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, host)
			}
		})
	}
}

func testBundle(t *testing.T) document.Bundle {
	t.Helper()
	bundle, err := document.NewBundleByPath("testdata")
	require.NoError(t, err)
	return bundle
}

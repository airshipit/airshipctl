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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/inventory/ifc"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
)

const (
	kind       = "BareMetalHost"
	bmhMaster0 = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-0
spec:
  online: true
  bootMACAddress: 00:3b:8b:0c:ec:8b
  bmc:
    address: redfish+http://nolocalhost:32201/redfish/v1/Systems/ephemeral
    credentialsName: master-0-bmc-secret
`
	master0BmcSec = `apiVersion: v1
kind: Secret
metadata:
  labels:
  name: master-0-bmc-secret
type: Opaque
data:
  username: YWRtaW4=
  password: cGFzc3dvcmQ=
`
	bmhMaster1 = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    host-group: "control-plane"
  name: master-1
spec:
  online: true
  bootMACAddress: 00:3b:8b:0c:ec:8b
  bmc:
    address: redfish+http://nolocalhost:8888/redfish/v1/Systems/node-master-1
    credentialsName: master-1-bmc-secret
`
	bmhMaster2 = `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  labels:
    host-group: "control-plane"
  name: master-2
spec:
  online: true
  bootMACAddress: 00:3b:8b:0c:ec:8b
  bmc:
    address: redfish+http://nolocalhost:8888/redfish/v1/Systems/node-master-2
    credentialsName: master-1-bmc-secret
`
	bmhNoCreds = `apiVersion: v1
kind: Secret
metadata:
  labels:
    airshipit.org/ephemeral-node: "true"
  name: master-1-bmc-secret
type: Opaque
data:
  username: YWRtaW4=
  password: cGFzc3dvcmQ=
`
)

func getMaster0Docs(t *testing.T) []document.Document {
	docCfgs := []string{
		bmhMaster0,
	}
	return buildTestDocs(t, docCfgs)
}

func getControlPlaneDocs(t *testing.T) []document.Document {
	docCfgs := []string{
		bmhMaster1,
		bmhMaster2,
	}
	return buildTestDocs(t, docCfgs)
}

func getNoCredsDocs(t *testing.T) []document.Document {
	docCfgs := []string{
		bmhNoCreds,
	}
	return buildTestDocs(t, docCfgs)
}

func getNoSuchHostDocs(t *testing.T) []document.Document {
	docCfgs := []string{}
	return buildTestDocs(t, docCfgs)
}

func buildTestDocs(t *testing.T, docCfgs []string) []document.Document {
	allDocs := make([]document.Document, len(docCfgs))
	for i, cfg := range docCfgs {
		doc, err := document.NewDocumentFromBytes([]byte(cfg))
		require.NoError(t, err)
		allDocs[i] = doc
	}
	return allDocs
}

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

	bundle := testSelectBundle(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mgmCfg := &config.ManagementConfiguration{Type: tt.remoteDriver}
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

	bundle := testSelectOneBundle(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mgmCfg := &config.ManagementConfiguration{Type: tt.remoteDriver}
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

func TestRunAction(t *testing.T) {
	tests := []struct {
		name, remoteDriver, expectedErr string
		operation                       ifc.BaremetalOperation

		selector ifc.BaremetalHostSelector
	}{
		{
			name:         "success return one host",
			remoteDriver: "redfish",
			operation:    ifc.BaremetalOperation("not supported"),
			selector:     (ifc.BaremetalHostSelector{}).ByName("master-0"),
			expectedErr:  "Baremetal operation not supported",
		},
		{
			name:         "success return one host",
			remoteDriver: "redfish",
			operation:    ifc.BaremetalOperationPowerOn,
			selector:     (ifc.BaremetalHostSelector{}).ByName("does not exist"),
			expectedErr:  "No baremetal hosts matched selector",
		},
		{
			name:         "success return one host",
			remoteDriver: "redfish",
			operation:    ifc.BaremetalOperationPowerOn,
			selector:     (ifc.BaremetalHostSelector{}).ByName("master-0"),
			expectedErr:  "HTTP request failed",
		},
	}

	bundle := testSelectBundle(t)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mgmCfg := config.ManagementConfiguration{Type: tt.remoteDriver}
			inventory := NewInventory(&mgmCfg, bundle)
			err := inventory.RunOperation(
				context.Background(),
				tt.operation,
				tt.selector,
				ifc.BaremetalBatchRunOptions{})
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAction(t *testing.T) {
	tests := []struct {
		name      string
		action    ifc.BaremetalOperation
		expectErr bool
	}{
		{
			name:   "poweron",
			action: ifc.BaremetalOperationPowerOn,
		},
		{
			name:   "poweroff",
			action: ifc.BaremetalOperationPowerOff,
		},
		{
			name:   "ejectvirtualmedia",
			action: ifc.BaremetalOperationEjectVirtualMedia,
		},
		{
			name:   "reboot",
			action: ifc.BaremetalOperationReboot,
		},
		{
			name:      "reboot",
			action:    ifc.BaremetalOperation("not supported"),
			expectErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actionFunc, err := action(context.Background(), tt.action)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// TODO inject fake host interface here to validate
				// that correct actions were selected
				assert.NotNil(t, actionFunc)
			}
		})
	}
}

func testSelectBundle(t *testing.T) document.Bundle {
	t.Helper()
	bundle := &testdoc.MockBundle{}
	secDoc, err := document.NewDocumentFromBytes([]byte(master0BmcSec))
	require.NoError(t, err)

	bundle.On("SelectOne", mock.Anything).
		Return(secDoc, nil)
	bundle.On("Select", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.Name == "master-0"
	})).Return(getMaster0Docs(t), nil)
	bundle.On("Select", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.LabelSelector == "host-group=control-plane"
	})).Return(getControlPlaneDocs(t), nil)
	bundle.On("Select", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.Name == "no-creds"
	})).Return(getNoCredsDocs(t), nil)
	bundle.On("Select", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.Name == "no such host"
	})).Return(getNoSuchHostDocs(t), nil)
	bundle.On("Select", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.Name == "does not exist"
	})).Return(getNoSuchHostDocs(t), nil)

	return bundle
}

func testSelectOneBundle(t *testing.T) document.Bundle {
	t.Helper()
	bundle := &testdoc.MockBundle{}
	secDoc, err := document.NewDocumentFromBytes([]byte(master0BmcSec))
	require.NoError(t, err)
	bmhMaster0Doc, err := document.NewDocumentFromBytes([]byte(bmhMaster0))
	require.NoError(t, err)

	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.Name == "master-0"
	})).Return(bmhMaster0Doc, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == "Secret" && selector.Name == "master-0-bmc-secret"
	})).Return(secDoc, nil)
	bundle.On("SelectOne", mock.MatchedBy(func(selector document.Selector) bool {
		return selector.Kind == kind && selector.LabelSelector == "host-group=control-plane"
	})).Return(nil, errors.New("found more than one document"))

	return bundle
}

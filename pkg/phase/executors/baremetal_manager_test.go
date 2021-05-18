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

package executors_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	inventoryifc "opendev.org/airship/airshipctl/pkg/inventory/ifc"
	"opendev.org/airship/airshipctl/pkg/k8s/utils"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
	testinventory "opendev.org/airship/airshipctl/testutil/inventory"
)

var bmhExecutorTemplate = `apiVersion: airshipit.org/v1alpha1
kind: BaremetalManager
metadata:
  name: RemoteDirectEphemeral
  labels:
    airshipit.org/deploy-k8s: "false"
spec:
  operation: "%s"
  hostSelector:
    name: node02
  operationOptions:
    remoteDirect:
      isoURL: %s`

func testBaremetalInventory() inventoryifc.Inventory {
	bmhi := &testinventory.MockBMHInventory{}
	bmhi.On("SelectOne", mock.Anything).Return()
	bi := &testinventory.MockInventory{}
	bi.On("BaremetalInventory").Return(bmhi, nil)
	return bi
}

func testBaremetalInventoryNoKustomization() inventoryifc.Inventory {
	bi := &testinventory.MockInventory{}
	bi.On("BaremetalInventory").
		Return(nil, errors.New("there is no kustomization.yaml"))
	return bi
}

func TestNewBMHExecutor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		execDoc := executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "reboot", "/home/iso-url"))
		executor, err := executors.NewBaremetalExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testBundleFactory(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, executor)
	})

	t.Run("error", func(t *testing.T) {
		exepectedErr := fmt.Errorf("ToAPI error")
		execDoc := &testdoc.MockDocument{
			MockToAPIObject: func() error { return exepectedErr },
		}
		executor, actualErr := executors.NewBaremetalExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testBundleFactory(),
		})
		assert.Equal(t, exepectedErr, actualErr)
		assert.Nil(t, executor)
	})
}

func TestBMHExecutorRun(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr string
		runOptions  ifc.RunOptions
		execDoc     document.Document
		inventory   inventoryifc.Inventory
	}{
		{
			name:        "error validate dry-run",
			expectedErr: "unknown action type",
			runOptions: ifc.RunOptions{
				DryRun: true,
				// any value but zero
				Timeout: 40,
			},
			execDoc:   executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "unknown", "")),
			inventory: testBaremetalInventory(),
		},
		{
			name: "success validate dry-run",
			runOptions: ifc.RunOptions{
				DryRun: true,
			},
			execDoc:   executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "remote-direct", "/some/url")),
			inventory: testBaremetalInventory(),
		},
		{
			name:        "error unknown action type",
			runOptions:  ifc.RunOptions{},
			execDoc:     executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "unknown", "")),
			expectedErr: "unknown action type",
			inventory:   testBaremetalInventory(),
		},
		{
			name:        "error no kustomization.yaml for inventory remote-direct",
			runOptions:  ifc.RunOptions{},
			execDoc:     executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "remote-direct", "")),
			expectedErr: "kustomization.yaml",
			inventory:   testBaremetalInventoryNoKustomization(),
		},
		{
			name:        "error no kustomization.yaml for inventory reboot",
			runOptions:  ifc.RunOptions{},
			execDoc:     executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "reboot", "")),
			expectedErr: "kustomization.yaml",
			inventory:   testBaremetalInventoryNoKustomization(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executor, err := executors.NewBaremetalExecutor(ifc.ExecutorConfig{
				ExecutorDocument: tt.execDoc,
				Inventory:        tt.inventory,
			})
			require.NoError(t, err)
			require.NotNil(t, executor)
			ch := make(chan events.Event)
			go func() {
				executor.Run(ch, tt.runOptions)
			}()
			processor := events.NewDefaultProcessor(utils.Streams())
			defer processor.Close()
			err = processor.Process(ch)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBMHValidate(t *testing.T) {
	tests := []struct {
		name        string
		expectedErr string
		execDoc     document.Document
	}{
		{
			name:        "error validate unknown action",
			expectedErr: "unknown action type",
			execDoc:     executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "unknown", "")),
		},
		{
			name:    "success validate remote-direct",
			execDoc: executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "remote-direct", "/some/url")),
		},
		{
			name:    "success validate reboot",
			execDoc: executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "reboot", "/some/url")),
		},
		{
			name:    "success validate power-off",
			execDoc: executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "power-off", "/some/url")),
		},
		{
			name:    "success validate power-on",
			execDoc: executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "power-on", "/some/url")),
		},
		{
			name:    "success validate eject-virtual-media",
			execDoc: executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "eject-virtual-media", "/some/url")),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executor, err := executors.NewBaremetalExecutor(ifc.ExecutorConfig{
				ExecutorDocument: tt.execDoc,
			})
			require.NoError(t, err)
			require.NotNil(t, executor)

			actualErr := executor.Validate()
			if tt.expectedErr != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, actualErr)
			}
		})
	}
}

// Dummy test to keep up with coverage, develop better testcases when render is implemented
func TestBMHManagerRender(t *testing.T) {
	execDoc := executorDoc(t, fmt.Sprintf(bmhExecutorTemplate, "reboot", "/home/iso-url"))
	executor, err := executors.NewBaremetalExecutor(ifc.ExecutorConfig{
		ExecutorDocument: execDoc,
	})
	require.NoError(t, err)
	require.NotNil(t, executor)

	err = executor.Render(bytes.NewBuffer([]byte{}), ifc.RenderOptions{})
	assert.NoError(t, err)
}

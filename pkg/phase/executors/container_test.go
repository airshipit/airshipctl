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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	containerExecutorDoc = `
  apiVersion: airshipit.org/v1alpha1
  kind: GenericContainer
  metadata:
    name: builder
    labels:
      airshipit.org/deploy-k8s: "false"
  spec:
    sinkOutputDir: "target/generator/results/generated"
    type: krm
    image: builder1
    mounts:
    - type: bind
      src: /home/ubuntu/mounts
      dst: /my-mounts
      rw: true
  config: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-srange-name
    data:
      cmd: encrypt
      unencrypted-regex: '^(kind|apiVersion|group|metadata)$'`
	singleExecutorBundlePath = "../../container/testdata/single"
)

func TestNewContainerExecutor(t *testing.T) {
	execDoc, err := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, err)

	t.Run("success new container executor", func(t *testing.T) {
		e, err := executors.NewContainerExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testBundleFactory(singleExecutorBundlePath),
			Helper:           makeDefaultHelper(t, "../../container/testdata"),
		})
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("error bundle factory", func(t *testing.T) {
		e, err := executors.NewContainerExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory: func() (document.Bundle, error) {
				return nil, fmt.Errorf("bundle error")
			},
			Helper: makeDefaultHelper(t, "../../container/testdata"),
		})
		assert.Error(t, err)
		assert.Nil(t, e)
	})
}

func TestGenericContainer(t *testing.T) {
	tests := []struct {
		name        string
		outputPath  string
		expectedErr string

		containerAPI   *v1alpha1.GenericContainer
		executorConfig ifc.ExecutorConfig
		runOptions     ifc.RunOptions
		clientFunc     container.ClientV1Alpha1FactoryFunc
	}{
		{
			name:        "error unknown container type",
			expectedErr: "uknown generic container type",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type: "unknown",
				},
			},
			clientFunc: container.NewClientV1Alpha1,
		},
		{
			name: "error kyaml cant parse config",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type: v1alpha1.GenericContainerTypeKrm,
				},
				Config: "~:~",
			},
			runOptions:  ifc.RunOptions{},
			expectedErr: "wrong Node Kind",
			clientFunc:  container.NewClientV1Alpha1,
		},
		{
			name:         "success dry run",
			containerAPI: &v1alpha1.GenericContainer{},
			runOptions:   ifc.RunOptions{DryRun: true},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b, err := document.NewBundleByPath(singleExecutorBundlePath)
			require.NoError(t, err)
			container := executors.ContainerExecutor{
				ResultsDir:     tt.outputPath,
				ExecutorBundle: b,
				Container:      tt.containerAPI,
				ClientFunc:     tt.clientFunc,
			}

			ch := make(chan events.Event)
			go container.Run(ch, tt.runOptions)

			var actualEvt []events.Event
			for evt := range ch {
				actualEvt = append(actualEvt, evt)
			}
			require.Greater(t, len(actualEvt), 0)

			if tt.expectedErr != "" {
				e := actualEvt[len(actualEvt)-1]
				require.Error(t, e.ErrorEvent.Error)
				assert.Contains(t, e.ErrorEvent.Error.Error(), tt.expectedErr)
			} else {
				e := actualEvt[len(actualEvt)-1]
				assert.NoError(t, e.ErrorEvent.Error)
				assert.Equal(t, e.Type, events.GenericContainerType)
				assert.Equal(t, e.GenericContainerEvent.Operation, events.GenericContainerStop)
			}
		})
	}
}

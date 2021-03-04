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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
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
			Helper:           makeDefaultHelper(t, "../../container/testdata", "metadata.yaml"),
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
			Helper: makeDefaultHelper(t, "../../container/testdata", "metadata.yaml"),
		})
		assert.Error(t, err)
		assert.Nil(t, e)
	})
}

func TestGenericContainer(t *testing.T) {
	tests := []struct {
		name         string
		outputPath   string
		expectedErr  string
		targetPath   string
		resultConfig string

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
			targetPath: singleExecutorBundlePath,
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
			targetPath:  singleExecutorBundlePath,
		},
		{
			name: "error no object referenced in config",
			containerAPI: &v1alpha1.GenericContainer{
				ConfigRef: &v1.ObjectReference{
					Kind: "no such kind",
					Name: "no such name",
				},
			},
			runOptions:  ifc.RunOptions{DryRun: true},
			expectedErr: "found no documents",
			targetPath:  singleExecutorBundlePath,
		},
		{
			name:         "success dry run",
			containerAPI: &v1alpha1.GenericContainer{},
			runOptions:   ifc.RunOptions{DryRun: true},
			targetPath:   singleExecutorBundlePath,
		},
		{
			name: "success referenced config present",
			containerAPI: &v1alpha1.GenericContainer{
				ConfigRef: &v1.ObjectReference{
					Kind:       "Secret",
					Name:       "test-script",
					APIVersion: "v1",
				},
			},
			runOptions: ifc.RunOptions{DryRun: true},
			targetPath: singleExecutorBundlePath,
			resultConfig: `apiVersion: v1
kind: Secret
metadata:
  name: test-script
stringData:
  script.sh: |
    #!/bin/sh
    echo WORKS! $var >&2
type: Opaque
`,
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
				Helper:         makeDefaultHelper(t, tt.targetPath, "../metadata.yaml"),
				Options: ifc.ExecutorConfig{
					ClusterName: "testCluster",
					KubeConfig: fakeKubeConfig{
						getFile: func() (string, kubeconfig.Cleanup, error) {
							return "testPath", func() {}, nil
						},
					},
				},
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
				assert.Equal(t, tt.resultConfig, container.Container.Config)
			}
		})
	}
}

func TestSetKubeConfig(t *testing.T) {
	getFileErr := fmt.Errorf("failed to get file")
	testCases := []struct {
		name        string
		opts        ifc.ExecutorConfig
		expectedErr error
	}{
		{
			name: "Set valid kubeconfig",
			opts: ifc.ExecutorConfig{
				ClusterName: "testCluster",
				KubeConfig: fakeKubeConfig{
					getFile: func() (string, kubeconfig.Cleanup, error) {
						return "testPath", func() {}, nil
					},
				},
			},
		},
		{
			name: "Failed to get kubeconfig file",
			opts: ifc.ExecutorConfig{
				ClusterName: "testCluster",
				KubeConfig: fakeKubeConfig{
					getFile: func() (string, kubeconfig.Cleanup, error) {
						return "", func() {}, getFileErr
					},
				},
			},
			expectedErr: getFileErr,
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			e := executors.ContainerExecutor{
				Options:   tt.opts,
				Container: &v1alpha1.GenericContainer{},
				Helper:    makeDefaultHelper(t, singleExecutorBundlePath, "../metadata.yaml"),
			}
			_, err := e.SetKubeConfig()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

type fakeKubeConfig struct {
	getFile func() (string, kubeconfig.Cleanup, error)
}

func (k fakeKubeConfig) GetFile() (string, kubeconfig.Cleanup, error) { return k.getFile() }
func (k fakeKubeConfig) Write(_ io.Writer) error                      { return nil }
func (k fakeKubeConfig) WriteFile(path string) error                  { return nil }
func (k fakeKubeConfig) WriteTempFile(dumpRoot string) (string, kubeconfig.Cleanup, error) {
	return k.getFile()
}

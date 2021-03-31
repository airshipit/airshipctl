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
	goerrors "errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase/errors"
	"opendev.org/airship/airshipctl/pkg/phase/executors"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
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
    - type: bind
      src: ~/mounts
      dst: /my-tilde-mount
      rw: true
  config: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: my-strange-name
    data:
      cmd: encrypt
      unencrypted-regex: '^(kind|apiVersion|group|metadata)$'`

	singleExecutorClusterMap = `apiVersion: airshipit.org/v1alpha1
kind: ClusterMap
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: main-map
map:
  testCluster: {}
`

	refConfig = `apiVersion: v1
kind: Secret
metadata:
  name: test-script
stringData:
  script.sh: |
    #!/bin/sh
    echo WORKS! $var >&2
type: Opaque
`
)

func testContainerBundleFactory() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		bundle := &testdoc.MockBundle{}
		return bundle, nil
	}
}

func testContainerBundleFactoryNoDocumentEntryPoint() document.BundleFactoryFunc {
	return func() (document.Bundle, error) {
		return nil, errors.ErrDocumentEntrypointNotDefined{}
	}
}

func testContainerBundle(t *testing.T, content string) document.Bundle {
	bundle := &testdoc.MockBundle{}
	writer := &bytes.Buffer{}
	bundle.On("Write", writer).
		Return(nil).
		Run(func(args mock.Arguments) {
			arg, ok := args.Get(0).(*bytes.Buffer)
			if ok {
				_, err := arg.Write([]byte(content))
				require.NoError(t, err)
			}
		})
	return bundle
}

func testContainerPhaseConfigBundleNoDocs() document.Bundle {
	bundle := &testdoc.MockBundle{}
	bundle.On("SelectOne", mock.Anything).
		Return(nil, goerrors.New("found no documents"))
	return bundle
}

func testContainerPhaseConfigBundleRefConfig(t *testing.T) document.Bundle {
	refConfigDoc, err := document.NewDocumentFromBytes([]byte(refConfig))
	require.NoError(t, err)
	bundle := &testdoc.MockBundle{}
	bundle.On("SelectOne", mock.Anything).
		Return(refConfigDoc, nil)
	return bundle
}

func TestNewContainerExecutor(t *testing.T) {
	execDoc, errDoc := document.NewDocumentFromBytes([]byte(containerExecutorDoc))
	require.NoError(t, errDoc)

	t.Run("success new container executor", func(t *testing.T) {
		e, err := executors.NewContainerExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testContainerBundleFactory(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})

	t.Run("error bundle factory", func(t *testing.T) {
		e, err := executors.NewContainerExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testdoc.ErrorBundleFactory,
		})
		assert.Error(t, err)
		assert.Nil(t, e)
	})

	t.Run("bundle factory - empty documentEntryPoint", func(t *testing.T) {
		e, err := executors.NewContainerExecutor(ifc.ExecutorConfig{
			ExecutorDocument: execDoc,
			BundleFactory:    testContainerBundleFactoryNoDocumentEntryPoint(),
		})
		assert.NoError(t, err)
		assert.NotNil(t, e)
	})
}

func TestGenericContainer(t *testing.T) {
	tests := []struct {
		name         string
		outputPath   string
		expectedErr  string
		resultConfig string

		containerAPI      *v1alpha1.GenericContainer
		executorConfig    ifc.ExecutorConfig
		runOptions        ifc.RunOptions
		clientFunc        container.ClientV1Alpha1FactoryFunc
		phaseConfigBundle document.Bundle
	}{
		{
			name:        "error unknown container type",
			expectedErr: "unknown generic container type",
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
			name: "error no object referenced in config",
			containerAPI: &v1alpha1.GenericContainer{
				ConfigRef: &v1.ObjectReference{
					Kind: "no such kind",
					Name: "no such name",
				},
			},
			runOptions:        ifc.RunOptions{DryRun: true},
			expectedErr:       "found no documents",
			phaseConfigBundle: testContainerPhaseConfigBundleNoDocs(),
		},
		{
			name:         "success dry run",
			containerAPI: &v1alpha1.GenericContainer{},
			runOptions:   ifc.RunOptions{DryRun: true},
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
			runOptions:        ifc.RunOptions{DryRun: true},
			resultConfig:      refConfig,
			phaseConfigBundle: testContainerPhaseConfigBundleRefConfig(t),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			executorBundle := testContainerBundle(t, "")

			containerExecutor := executors.ContainerExecutor{
				ResultsDir:     tt.outputPath,
				ExecutorBundle: executorBundle,
				Container:      tt.containerAPI,
				ClientFunc:     tt.clientFunc,
				Options: ifc.ExecutorConfig{
					ClusterName: "testCluster",
					KubeConfig: fakeKubeConfig{
						getFile: func() (string, kubeconfig.Cleanup, error) {
							return "testPath", func() {}, nil
						},
					},
					ClusterMap:        testClusterMap(t),
					PhaseConfigBundle: tt.phaseConfigBundle,
				},
			}

			ch := make(chan events.Event)
			go containerExecutor.Run(ch, tt.runOptions)

			actualEvt := make([]events.Event, 0)
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
				assert.Equal(t, tt.resultConfig, containerExecutor.Container.Config)
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
				ClusterMap: testClusterMap(t),
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
				ClusterMap: testClusterMap(t),
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
			}
			_, err := e.SetKubeConfig()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestContainerRender(t *testing.T) {
	testCases := []struct {
		name        string
		bundleData  []byte
		opts        ifc.RenderOptions
		expectedErr error
		expectedOut string
	}{
		{
			name:        "empty bundle",
			bundleData:  []byte{},
			opts:        ifc.RenderOptions{},
			expectedOut: "",
			expectedErr: nil,
		},
		{
			name: "valid bundle",
			bundleData: []byte(`apiVersion: unittest.org/v1alpha1
kind: Test
metadata:
  name: TestName`),
			opts: ifc.RenderOptions{FilterSelector: document.NewSelector().ByKind("Test").ByName("TestName")},
			expectedOut: `---
apiVersion: unittest.org/v1alpha1
kind: Test
metadata:
  name: TestName
...
`,
			expectedErr: nil,
		},
	}

	buf := &bytes.Buffer{}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := document.NewBundleFromBytes(tt.bundleData)
			require.NoError(t, err)
			e := executors.ContainerExecutor{
				ExecutorBundle: bundle,
			}
			err = e.Render(buf, tt.opts)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedOut, buf.String())
			buf.Reset()
		})
	}
}

type fakeKubeConfig struct {
	getFile func() (string, kubeconfig.Cleanup, error)
}

func (k fakeKubeConfig) GetFile() (string, kubeconfig.Cleanup, error)        { return k.getFile() }
func (k fakeKubeConfig) Write(_ io.Writer) error                             { return nil }
func (k fakeKubeConfig) WriteFile(_ string, _ kubeconfig.WriteOptions) error { return nil }
func (k fakeKubeConfig) WriteTempFile(_ string) (string, kubeconfig.Cleanup, error) {
	return k.getFile()
}

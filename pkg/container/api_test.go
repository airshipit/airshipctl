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

package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
	"opendev.org/airship/airshipctl/pkg/util"
)

func bundlePathToInput(t *testing.T, bundlePath string) io.Reader {
	t.Helper()
	bundle, err := document.NewBundleByPath(bundlePath)
	require.NoError(t, err)
	buf := bytes.NewBuffer([]byte{})
	err = bundle.Write(buf)
	require.NoError(t, err)
	return buf
}

func TestGenericContainer(t *testing.T) {
	// TODO add testcase were we mock KRM call, and make sure we put correct input into it
	tests := []struct {
		name            string
		inputBundlePath string
		outputPath      string
		expectedErr     string

		output         io.Writer
		containerAPI   *v1alpha1.GenericContainer
		execFunc       containerFunc
		executorConfig ifc.ExecutorConfig
	}{
		{
			name:        "error unknown container type",
			expectedErr: "unknown generic container type",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type: "unknown",
				},
			},
			execFunc: NewContainer,
		},
		{
			name: "error kyaml can't parse config",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type: v1alpha1.GenericContainerTypeKrm,
				},
				Config: "~:~",
			},
			execFunc:    NewContainer,
			expectedErr: "wrong Node Kind",
		},
		{
			name: "error runFns execute",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type: v1alpha1.GenericContainerTypeKrm,
					StorageMounts: []v1alpha1.StorageMount{
						{
							MountType: "bind",
							Src:       "test",
							DstPath:   "/mount",
						},
					},
				},
				Config: `kind: ConfigMap`,
			},
			execFunc:    NewContainer,
			expectedErr: "no such file or directory",
			outputPath:  "directory doesn't exist",
		},
		{
			name:       "error output directory does not exist",
			outputPath: "doesn't exist",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type:  v1alpha1.GenericContainerTypeAirship,
					Image: "some image",
					StorageMounts: []v1alpha1.StorageMount{
						{
							MountType: "bind",
							Src:       "test",
							DstPath:   "/mount",
						},
					},
				},
				Config: `kind: ConfigMap`,
			},
			expectedErr: "no such file or directory",
			execFunc: func(ctx context.Context, driver, url string) (Container, error) {
				return getDockerContainerMock(mockDockerClient{
					containerAttach: func() (types.HijackedResponse, error) {
						conn := types.HijackedResponse{
							Conn: mockConn{WData: make([]byte, len([]byte("foo: bar")))},
						}
						return conn, nil
					},
					imageList: func() ([]types.ImageSummary, error) {
						return []types.ImageSummary{{ID: "imgid"}}, nil
					},
					imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
						return types.ImageInspect{
							Config: &container.Config{
								Cmd: []string{"testCmd"},
							},
						}, nil, nil
					},
				}), nil
			},
		},
		{
			name:        "error airship container writeLogs",
			expectedErr: "container logs error",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type:  v1alpha1.GenericContainerTypeAirship,
					Image: "some image",
					StorageMounts: []v1alpha1.StorageMount{
						{
							MountType: "bind",
							Src:       "test",
							DstPath:   "/mount",
						},
						{
							MountType: "bind",
							Src:       "~/test",
							DstPath:   "/mnt",
						},
					},
				},
				Config: `kind: ConfigMap`,
			},
			execFunc: func(ctx context.Context, driver, url string) (Container, error) {
				return getDockerContainerMock(mockDockerClient{
					containerAttach: func() (types.HijackedResponse, error) {
						conn := types.HijackedResponse{
							Conn: mockConn{WData: make([]byte, len([]byte("foo: bar")))},
						}
						return conn, nil
					},
					imageList: func() ([]types.ImageSummary, error) {
						return []types.ImageSummary{{ID: "imgid"}}, nil
					},
					imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
						return types.ImageInspect{
							Config: &container.Config{
								Cmd: []string{"testCmd"},
							},
						}, nil, nil
					},
					containerLogs: func() (io.ReadCloser, error) {
						return nil, fmt.Errorf("container logs error")
					},
				}), nil
			},
		},
		{
			name: "basic success airship container",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type:  v1alpha1.GenericContainerTypeAirship,
					Image: "some image",
					StorageMounts: []v1alpha1.StorageMount{
						{
							MountType: "bind",
							Src:       "test",
							DstPath:   "/mount",
						},
						{
							MountType: "bind",
							Src:       "~/test",
							DstPath:   "/mnt",
						},
					},
				},
				Config: `kind: ConfigMap`,
			},
			execFunc: func(ctx context.Context, driver, url string) (Container, error) {
				return getDockerContainerMock(mockDockerClient{
					containerAttach: func() (types.HijackedResponse, error) {
						conn := types.HijackedResponse{
							Conn: mockConn{WData: make([]byte, len([]byte("foo: bar")))},
						}
						return conn, nil
					},
					imageList: func() ([]types.ImageSummary, error) {
						return []types.ImageSummary{{ID: "imgid"}}, nil
					},
					imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
						return types.ImageInspect{
							Config: &container.Config{
								Cmd: []string{"testCmd"},
							},
						}, nil, nil
					},
				}), nil
			},
		},
		{
			name: "basic success airship success written to provided output Writer",
			containerAPI: &v1alpha1.GenericContainer{
				Spec: v1alpha1.GenericContainerSpec{
					Type:  v1alpha1.GenericContainerTypeAirship,
					Image: "some image",
					StorageMounts: []v1alpha1.StorageMount{
						{
							MountType: "bind",
							Src:       "test",
							DstPath:   "/mount",
						},
					},
				},
				Config: `kind: ConfigMap`,
			},
			execFunc: func(ctx context.Context, driver, url string) (Container, error) {
				return getDockerContainerMock(mockDockerClient{
					containerAttach: func() (types.HijackedResponse, error) {
						conn := types.HijackedResponse{
							Conn: mockConn{WData: make([]byte, len([]byte("foo: bar")))},
						}
						return conn, nil
					},
					imageList: func() ([]types.ImageSummary, error) {
						return []types.ImageSummary{{ID: "imgid"}}, nil
					},
					imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
						return types.ImageInspect{
							Config: &container.Config{},
						}, nil, nil
					},
				}), nil
			},
			output: ioutil.Discard,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			input := bundlePathToInput(t, "testdata/single")
			client := &clientV1Alpha1{
				input:         input,
				resultsDir:    tt.outputPath,
				output:        tt.output,
				conf:          tt.containerAPI,
				containerFunc: tt.execFunc,
			}

			err := client.Run()

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Dummy test to keep up with coverage.
func TestNewClientV1alpha1(t *testing.T) {
	client := NewClientV1Alpha1("", nil, nil, v1alpha1.DefaultGenericContainer(), "")
	require.NotNil(t, client)
}

func TestExpandSourceMounts(t *testing.T) {
	tests := []struct {
		name           string
		inputMounts    []v1alpha1.StorageMount
		targetPath     string
		expectedMounts []v1alpha1.StorageMount
	}{
		{
			name:           "absolute path",
			inputMounts:    []v1alpha1.StorageMount{{Src: "/test"}},
			targetPath:     "/target-path",
			expectedMounts: []v1alpha1.StorageMount{{Src: "/test"}},
		},
		{
			name:           "tilde path",
			inputMounts:    []v1alpha1.StorageMount{{Src: "~/test"}},
			targetPath:     "/target-path",
			expectedMounts: []v1alpha1.StorageMount{{Src: util.ExpandTilde("~/test")}},
		},
		{
			name:           "relative path",
			inputMounts:    []v1alpha1.StorageMount{{Src: "test"}},
			targetPath:     "/target-path",
			expectedMounts: []v1alpha1.StorageMount{{Src: filepath.Join("/target-path", "test")}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ExpandSourceMounts(tt.inputMounts, tt.targetPath)
			require.Equal(t, tt.expectedMounts, tt.inputMounts)
		})
	}
}

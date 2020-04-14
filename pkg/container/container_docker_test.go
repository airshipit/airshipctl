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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
)

type mockConn struct {
	WData []byte
}

func (mc mockConn) Read(b []byte) (n int, err error) { return len(b), nil }
func (mc mockConn) Write(b []byte) (n int, err error) {
	copy(mc.WData, b)
	return len(b), nil
}
func (mc mockConn) Close() error                     { return nil }
func (mc mockConn) LocalAddr() net.Addr              { return nil }
func (mc mockConn) RemoteAddr() net.Addr             { return nil }
func (mc mockConn) SetDeadline(time.Time) error      { return nil }
func (mc mockConn) SetReadDeadline(time.Time) error  { return nil }
func (mc mockConn) SetWriteDeadline(time.Time) error { return nil }

type mockDockerClient struct {
	imageInspectWithRaw func() (types.ImageInspect, []byte, error)
	imageList           func() ([]types.ImageSummary, error)
	containerAttach     func() (types.HijackedResponse, error)
	imagePull           func() (io.ReadCloser, error)
	containerStart      func() error
	containerWait       func() (<-chan container.ContainerWaitOKBody, <-chan error)
	containerLogs       func() (io.ReadCloser, error)
}

func (mdc *mockDockerClient) ImageInspectWithRaw(context.Context, string) (types.ImageInspect, []byte, error) {
	return mdc.imageInspectWithRaw()
}
func (mdc *mockDockerClient) ImageList(context.Context, types.ImageListOptions) ([]types.ImageSummary, error) {
	return mdc.imageList()
}
func (mdc *mockDockerClient) ImagePull(
	context.Context,
	string,
	types.ImagePullOptions,
) (io.ReadCloser, error) {
	return mdc.imagePull()
}
func (mdc *mockDockerClient) ContainerCreate(
	context.Context,
	*container.Config,
	*container.HostConfig,
	*network.NetworkingConfig,
	string,
) (container.ContainerCreateCreatedBody, error) {
	return container.ContainerCreateCreatedBody{ID: "testID"}, nil
}
func (mdc *mockDockerClient) ContainerAttach(
	context.Context,
	string,
	types.ContainerAttachOptions,
) (types.HijackedResponse, error) {
	return mdc.containerAttach()
}
func (mdc *mockDockerClient) ContainerStart(context.Context, string, types.ContainerStartOptions) error {
	if mdc.containerStart != nil {
		return mdc.containerStart()
	}
	return nil
}
func (mdc *mockDockerClient) ContainerWait(
	context.Context,
	string,
	container.WaitCondition,
) (<-chan container.ContainerWaitOKBody, <-chan error) {
	if mdc.containerWait != nil {
		return mdc.containerWait()
	}

	resC := make(chan container.ContainerWaitOKBody)
	go func() {
		resC <- container.ContainerWaitOKBody{StatusCode: 0}
	}()
	return resC, nil
}
func (mdc *mockDockerClient) ContainerLogs(context.Context, string, types.ContainerLogsOptions) (io.ReadCloser, error) {
	if mdc.containerLogs != nil {
		return mdc.containerLogs()
	}
	return ioutil.NopCloser(strings.NewReader("")), nil
}

func (mdc *mockDockerClient) ContainerRemove(context.Context, string, types.ContainerRemoveOptions) error {
	return nil
}

func getDockerContainerMock(mdc mockDockerClient) *DockerContainer {
	ctx := context.Background()
	cnt := &DockerContainer{
		dockerClient: &mdc,
		ctx:          &ctx,
	}
	return cnt
}

func TestGetCmd(t *testing.T) {
	testError := fmt.Errorf("inspect error")
	tests := []struct {
		cmd              []string
		mockDockerClient mockDockerClient
		expectedResult   []string
		expectedErr      error
	}{
		{
			cmd:              []string{"testCmd"},
			mockDockerClient: mockDockerClient{},
			expectedResult:   []string{"testCmd"},
			expectedErr:      nil,
		},
		{
			cmd: []string{},
			mockDockerClient: mockDockerClient{
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
			},
			expectedResult: []string{"testCmd"},
			expectedErr:    nil,
		},
		{
			cmd: []string{},
			mockDockerClient: mockDockerClient{
				imageList: func() ([]types.ImageSummary, error) {
					return []types.ImageSummary{{ID: "imgid"}}, nil
				},
				imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
					return types.ImageInspect{}, nil, testError
				},
			},
			expectedResult: nil,
			expectedErr:    testError,
		},
	}

	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualRes, actualErr := cnt.getCmd(tt.cmd)

		assert.Equal(t, tt.expectedErr, actualErr)
		assert.Equal(t, tt.expectedResult, actualRes)
	}
}

func TestGetImageId(t *testing.T) {
	testError := fmt.Errorf("img list error")
	tests := []struct {
		url              string
		mockDockerClient mockDockerClient
		expectedResult   string
		expectedErr      error
	}{
		{
			url: "test:latest",
			mockDockerClient: mockDockerClient{
				imageList: func() ([]types.ImageSummary, error) {
					return nil, testError
				},
			},
			expectedResult: "",
			expectedErr:    testError,
		},
		{
			url: "test:latest",
			mockDockerClient: mockDockerClient{
				imageList: func() ([]types.ImageSummary, error) {
					return []types.ImageSummary{}, nil
				},
			},
			expectedResult: "",
			expectedErr:    ErrEmptyImageList{},
		},
	}
	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualRes, actualErr := cnt.getImageID(tt.url)

		assert.Equal(t, tt.expectedErr, actualErr)
		assert.Equal(t, tt.expectedResult, actualRes)
	}
}

func TestImagePull(t *testing.T) {
	testError := fmt.Errorf("image pull rror")
	tests := []struct {
		mockDockerClient mockDockerClient
		expectedErr      error
	}{
		{
			mockDockerClient: mockDockerClient{
				imagePull: func() (io.ReadCloser, error) {
					return ioutil.NopCloser(strings.NewReader("test")), nil
				},
				imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
					return types.ImageInspect{}, nil, testError
				},
			},
			expectedErr: nil,
		},
		{
			mockDockerClient: mockDockerClient{
				imagePull: func() (io.ReadCloser, error) {
					return nil, testError
				},
				imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
					return types.ImageInspect{}, nil, testError
				},
			},
			expectedErr: testError,
		},
	}
	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualErr := cnt.ImagePull()

		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

func TestGetId(t *testing.T) {
	cnt := getDockerContainerMock(mockDockerClient{})
	err := cnt.RunCommand([]string{"testCmd"}, nil, nil, []string{}, false)
	require.NoError(t, err)
	actualID := cnt.GetID()

	assert.Equal(t, "testID", actualID)
}

func TestRunCommand(t *testing.T) {
	imageListError := fmt.Errorf("image list error")
	attachError := fmt.Errorf("attach error")
	containerStartError := fmt.Errorf("container start error")
	containerWaitError := fmt.Errorf("container wait error")
	tests := []struct {
		cmd              []string
		containerInput   io.Reader
		volumeMounts     []string
		debug            bool
		mockDockerClient mockDockerClient
		expectedErr      error
		assertF          func(*testing.T)
	}{
		{
			cmd:              []string{"testCmd"},
			containerInput:   nil,
			volumeMounts:     nil,
			debug:            false,
			mockDockerClient: mockDockerClient{},
			expectedErr:      nil,
			assertF:          func(t *testing.T) {},
		},
		{
			cmd:            []string{},
			containerInput: nil,
			volumeMounts:   nil,
			debug:          false,
			mockDockerClient: mockDockerClient{
				imageList: func() ([]types.ImageSummary, error) {
					return nil, imageListError
				},
			},
			expectedErr: imageListError,
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: strings.NewReader("testInput"),
			volumeMounts:   nil,
			debug:          false,
			mockDockerClient: mockDockerClient{
				containerAttach: func() (types.HijackedResponse, error) {
					conn := types.HijackedResponse{
						Conn: mockConn{WData: make([]byte, len([]byte("testInput")))},
					}
					return conn, nil
				},
			},
			expectedErr: nil,
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: strings.NewReader("testInput"),
			volumeMounts:   nil,
			debug:          false,
			mockDockerClient: mockDockerClient{
				containerAttach: func() (types.HijackedResponse, error) {
					return types.HijackedResponse{}, attachError
				},
			},
			expectedErr: attachError,
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: nil,
			volumeMounts:   nil,
			debug:          false,
			mockDockerClient: mockDockerClient{
				containerStart: func() error {
					return containerStartError
				},
			},
			expectedErr: containerStartError,
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: nil,
			volumeMounts:   nil,
			debug:          false,
			mockDockerClient: mockDockerClient{
				containerWait: func() (<-chan container.ContainerWaitOKBody, <-chan error) {
					errC := make(chan error)
					go func() {
						errC <- containerWaitError
					}()
					return nil, errC
				},
			},
			expectedErr: containerWaitError,
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: nil,
			volumeMounts:   nil,
			debug:          true,
			mockDockerClient: mockDockerClient{
				containerWait: func() (<-chan container.ContainerWaitOKBody, <-chan error) {
					resC := make(chan container.ContainerWaitOKBody)
					go func() {
						resC <- container.ContainerWaitOKBody{StatusCode: 1}
					}()
					return resC, nil
				},
			},
			expectedErr: ErrRunContainerCommand{Cmd: "docker logs testID"},
			assertF:     func(t *testing.T) {},
		},
		{
			cmd:            []string{"testCmd"},
			containerInput: nil,
			volumeMounts:   nil,
			debug:          true,
			mockDockerClient: mockDockerClient{
				containerWait: func() (<-chan container.ContainerWaitOKBody, <-chan error) {
					resC := make(chan container.ContainerWaitOKBody)
					go func() {
						resC <- container.ContainerWaitOKBody{StatusCode: 1}
					}()
					return resC, nil
				},
				containerLogs: func() (io.ReadCloser, error) {
					return nil, fmt.Errorf("logs error")
				},
			},
			expectedErr: ErrRunContainerCommand{Cmd: "docker logs testID"},
			assertF:     func(t *testing.T) {},
		},
	}
	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualErr := cnt.RunCommand(tt.cmd, tt.containerInput, tt.volumeMounts, []string{}, tt.debug)

		assert.Equal(t, tt.expectedErr, actualErr)

		tt.assertF(t)
	}
}

func TestRunCommandOutput(t *testing.T) {
	testError := fmt.Errorf("img list error")
	tests := []struct {
		cmd              []string
		containerInput   io.Reader
		volumeMounts     []string
		mockDockerClient mockDockerClient
		expectedResult   string
		expectedErr      error
	}{
		{
			cmd:              []string{"testCmd"},
			containerInput:   nil,
			volumeMounts:     nil,
			mockDockerClient: mockDockerClient{},
			expectedResult:   "",
			expectedErr:      nil,
		},
		{
			cmd:            []string{},
			containerInput: nil,
			volumeMounts:   nil,
			mockDockerClient: mockDockerClient{
				imageList: func() ([]types.ImageSummary, error) {
					return nil, testError
				},
			},
			expectedResult: "",
			expectedErr:    testError,
		},
	}
	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualRes, actualErr := cnt.RunCommandOutput(tt.cmd, tt.containerInput, tt.volumeMounts, []string{})

		assert.Equal(t, tt.expectedErr, actualErr)

		var actualResBytes []byte
		if actualRes != nil {
			var err error
			actualResBytes, err = ioutil.ReadAll(actualRes)
			require.NoError(t, err)
		} else {
			actualResBytes = []byte{}
		}

		assert.Equal(t, tt.expectedResult, string(actualResBytes))
	}
}

func TestNewDockerContainer(t *testing.T) {
	testError := fmt.Errorf("image pull error")
	type resultStruct struct {
		tag      string
		imageURL string
		id       string
	}

	tests := []struct {
		url            string
		ctx            context.Context
		cli            mockDockerClient
		expectedErr    error
		expectedResult resultStruct
	}{
		{
			url: "testPrefix/testImage:testTag",
			ctx: context.Background(),
			cli: mockDockerClient{
				imagePull: func() (io.ReadCloser, error) {
					return ioutil.NopCloser(strings.NewReader("test")), nil
				},
				imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
					return types.ImageInspect{}, nil, testError
				},
			},
			expectedErr: nil,
			expectedResult: resultStruct{
				tag:      "testTag",
				imageURL: "testPrefix/testImage:testTag",
				id:       "",
			},
		},
		{
			url: "testPrefix/testImage:testTag",
			ctx: context.Background(),
			cli: mockDockerClient{
				imagePull: func() (io.ReadCloser, error) {
					return nil, testError
				},
				imageInspectWithRaw: func() (types.ImageInspect, []byte, error) {
					return types.ImageInspect{}, nil, testError
				},
			},
			expectedErr:    testError,
			expectedResult: resultStruct{},
		},
	}
	for _, tt := range tests {
		actualRes, actualErr := NewDockerContainer(&(tt.ctx), tt.url, &(tt.cli))

		assert.Equal(t, tt.expectedErr, actualErr)

		var actualResStruct resultStruct
		if actualRes == nil {
			actualResStruct = resultStruct{}
		} else {
			actualResStruct = resultStruct{
				tag:      actualRes.tag,
				imageURL: actualRes.imageURL,
				id:       actualRes.id,
			}
		}
		assert.Equal(t, tt.expectedResult, actualResStruct)
	}
}

func TestRmContainer(t *testing.T) {
	tests := []struct {
		mockDockerClient mockDockerClient
		expectedErr      error
	}{
		{
			mockDockerClient: mockDockerClient{},
			expectedErr:      nil,
		},
	}

	for _, tt := range tests {
		cnt := getDockerContainerMock(tt.mockDockerClient)
		actualErr := cnt.RmContainer()
		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

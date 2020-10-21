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

package isogen

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	api "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/testutil"
)

type mockContainer struct {
	imagePull         func() error
	runCommand        func() error
	getContainerLogs  func() (io.ReadCloser, error)
	rmContainer       func() error
	getID             func() string
	waitUntilFinished func() error
	inspectContainer  func() (container.State, error)
}

func (mc *mockContainer) ImagePull() error {
	return mc.imagePull()
}

func (mc *mockContainer) RunCommand([]string, io.Reader, []string, []string) error {
	return mc.runCommand()
}

func (mc *mockContainer) GetContainerLogs() (io.ReadCloser, error) {
	return mc.getContainerLogs()
}

func (mc *mockContainer) RmContainer() error {
	return mc.rmContainer()
}

func (mc *mockContainer) GetID() string {
	return mc.getID()
}

func (mc *mockContainer) WaitUntilFinished() error {
	return mc.waitUntilFinished()
}

func (mc *mockContainer) InspectContainer() (container.State, error) {
	return mc.inspectContainer()
}

const testID = "TESTID"

func TestBootstrapIso(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata/primary/site/test-site/ephemeral/bootstrap")
	require.NoError(t, err, "Building Bundle Failed")

	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	volBind := tempVol + ":/dst"
	testErr := fmt.Errorf("TestErr")
	testCfg := &api.ImageConfiguration{
		Container: &api.Container{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Builder: &api.Builder{
			UserDataFileName:      "user-data",
			NetworkConfigFileName: "net-conf",
		},
	}
	testDoc := &MockDocument{
		MockAsYAML: func() ([]byte, error) { return []byte("TESTDOC"), nil },
	}
	testBuilder := &mockContainer{
		runCommand:  func() error { return nil },
		getID:       func() string { return testID },
		rmContainer: func() error { return nil },
	}

	expOut := []string{
		"Creating cloud-init for ephemeral K8s",
		fmt.Sprintf("Running default container command. Mounted dir: [%s]", volBind),
		"ISO successfully built.",
		"Debug flag is set. Container TESTID stopped but not deleted.",
		"Removing container.",
	}

	tests := []struct {
		builder     *mockContainer
		cfg         *api.ImageConfiguration
		doc         *MockDocument
		debug       bool
		expectedOut []string
		expectedErr error
	}{
		{
			builder: &mockContainer{
				runCommand:        func() error { return testErr },
				waitUntilFinished: func() error { return nil },
				rmContainer:       func() error { return nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1]},
			expectedErr: testErr,
		},
		{
			builder: &mockContainer{
				runCommand:        func() error { return nil },
				getID:             func() string { return "TESTID" },
				waitUntilFinished: func() error { return nil },
				rmContainer:       func() error { return nil },
				getContainerLogs:  func() (io.ReadCloser, error) { return ioutil.NopCloser(strings.NewReader("")), nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       true,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[3]},
			expectedErr: nil,
		},
		{
			builder: &mockContainer{
				runCommand:        func() error { return nil },
				getID:             func() string { return "TESTID" },
				rmContainer:       func() error { return testErr },
				waitUntilFinished: func() error { return nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[4]},
			expectedErr: testErr,
		},
		{
			builder: testBuilder,
			cfg:     testCfg,
			doc: &MockDocument{
				MockAsYAML: func() ([]byte, error) { return nil, testErr },
			},
			debug:       false,
			expectedOut: []string{expOut[0]},
			expectedErr: testErr,
		},
	}

	for _, tt := range tests {
		outBuf := &bytes.Buffer{}
		log.Init(tt.debug, outBuf)
		bootstrapOpts := BootstrapIsoOptions{
			docBundle: bundle,
			builder:   tt.builder,
			doc:       tt.doc,
			cfg:       tt.cfg,
			debug:     tt.debug,
		}
		actualErr := bootstrapOpts.createBootstrapIso()
		actualOut := outBuf.String()

		for _, line := range tt.expectedOut {
			assert.True(t, strings.Contains(actualOut, line))
		}

		assert.Equal(t, tt.expectedErr, actualErr)
	}
}

func TestVerifyInputs(t *testing.T) {
	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	tests := []struct {
		name        string
		cfg         *api.ImageConfiguration
		args        []string
		expectedErr error
	}{
		{
			name: "missing-container-field",
			cfg: &api.ImageConfiguration{
				Container: &api.Container{},
			},
			expectedErr: config.ErrMissingConfig{
				What: "Must specify volume bind for ISO builder container",
			},
		},
		{
			name: "missing-filenames",
			cfg: &api.ImageConfiguration{
				Container: &api.Container{
					Volume: tempVol + ":/dst",
				},
				Builder: &api.Builder{},
			},
			expectedErr: config.ErrMissingConfig{
				What: "UserDataFileName or NetworkConfigFileName are not specified in ISO builder config",
			},
		},
		{
			name: "invalid-host-path",
			cfg: &api.ImageConfiguration{
				Container: &api.Container{
					Volume: tempVol + ":/dst:/dst1",
				},
				Builder: &api.Builder{
					UserDataFileName:      "user-data",
					NetworkConfigFileName: "net-conf",
				},
			},
			expectedErr: config.ErrInvalidConfig{
				What: "Bad container volume format. Use hostPath:contPath",
			},
		},
		{
			name: "success",
			cfg: &api.ImageConfiguration{
				Container: &api.Container{
					Volume: tempVol,
				},
				Builder: &api.Builder{
					UserDataFileName:      "user-data",
					NetworkConfigFileName: "net-conf",
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			actualErr := verifyInputs(tt.cfg)
			assert.Equal(subTest, tt.expectedErr, actualErr)
		})
	}
}

func TestShowProgress(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "Process-debian-based-logs",
			input:  "testdata/debian-container-logs",
			output: "testdata/pb-output-debian",
		},
	}

	for _, tt := range tests {
		tt := tt

		testInput, err := ioutil.ReadFile(tt.input)
		require.NoError(t, err)
		reader := ioutil.NopCloser(bytes.NewReader(testInput))
		writer := bytes.NewBuffer(nil)
		err = showProgress(reader, writer)
		require.NoError(t, err)
		assert.Contains(t, writer.String(), "Completed")
	}
}

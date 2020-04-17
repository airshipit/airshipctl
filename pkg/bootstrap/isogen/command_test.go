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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/testutil"
)

type mockContainer struct {
	imagePull        func() error
	runCommand       func() error
	runCommandOutput func() (io.ReadCloser, error)
	rmContainer      func() error
	getID            func() string
}

func (mc *mockContainer) ImagePull() error {
	return mc.imagePull()
}

func (mc *mockContainer) RunCommand([]string, io.Reader, []string, []string, bool) error {
	return mc.runCommand()
}

func (mc *mockContainer) RunCommandOutput([]string, io.Reader, []string, []string) (io.ReadCloser, error) {
	return mc.runCommandOutput()
}

func (mc *mockContainer) RmContainer() error {
	return mc.rmContainer()
}

func (mc *mockContainer) GetID() string {
	return mc.getID()
}

func TestBootstrapIso(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata/primary/site/test-site/ephemeral/bootstrap")
	require.NoError(t, err, "Building Bundle Failed")

	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	volBind := tempVol + ":/dst"
	testErr := fmt.Errorf("TestErr")
	testCfg := &config.Bootstrap{
		Container: &config.Container{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Builder: &config.Builder{
			UserDataFileName:      "user-data",
			NetworkConfigFileName: "net-conf",
		},
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
		cfg         *config.Bootstrap
		debug       bool
		expectedOut []string
		expectedErr error
	}{
		{
			builder: &mockContainer{
				runCommand: func() error { return testErr },
			},
			cfg:         testCfg,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1]},
			expectedErr: testErr,
		},
		{
			builder: &mockContainer{
				runCommand: func() error { return nil },
				getID:      func() string { return "TESTID" },
			},
			cfg:         testCfg,
			debug:       true,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[3]},
			expectedErr: nil,
		},
		{
			builder: &mockContainer{
				runCommand:  func() error { return nil },
				getID:       func() string { return "TESTID" },
				rmContainer: func() error { return testErr },
			},
			cfg:         testCfg,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[4]},
			expectedErr: testErr,
		},
	}

	for _, tt := range tests {
		outBuf := &bytes.Buffer{}
		log.Init(tt.debug, outBuf)
		actualErr := generateBootstrapIso(bundle, tt.builder, tt.cfg, tt.debug)
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
		cfg         *config.Bootstrap
		args        []string
		expectedErr error
	}{
		{
			name: "missing-container-field",
			cfg: &config.Bootstrap{
				Container: &config.Container{},
			},
			expectedErr: config.ErrMissingConfig{
				What: "Must specify volume bind for ISO builder container",
			},
		},
		{
			name: "missing-filenames",
			cfg: &config.Bootstrap{
				Container: &config.Container{
					Volume: tempVol + ":/dst",
				},
				Builder: &config.Builder{},
			},
			expectedErr: config.ErrMissingConfig{
				What: "UserDataFileName or NetworkConfigFileName are not specified in ISO builder config",
			},
		},
		{
			name: "invalid-host-path",
			cfg: &config.Bootstrap{
				Container: &config.Container{
					Volume: tempVol + ":/dst:/dst1",
				},
				Builder: &config.Builder{
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
			cfg: &config.Bootstrap{
				Container: &config.Container{
					Volume: tempVol,
				},
				Builder: &config.Builder{
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

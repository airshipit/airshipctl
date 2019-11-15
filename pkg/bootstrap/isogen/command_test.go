package isogen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/testutil"
)

type mockContainer struct {
	imagePull        func() error
	runCommand       func() error
	runCommandOutput func() (io.ReadCloser, error)
	rmContainer      func() error
	getId            func() string
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

func (mc *mockContainer) GetId() string {
	return mc.getId()
}

func TestBootstrapIso(t *testing.T) {
	fSys := testutil.SetupTestFs(t, "testdata")
	bundle, err := document.NewBundle(fSys, "/", "/")
	require.NoError(t, err, "Building Bundle Failed")

	volBind := "/tmp:/dst"
	testErr := fmt.Errorf("TestErr")
	testCfg := &Config{
		Container: Container{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Builder: Builder{
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
		cfg         *Config
		debug       bool
		expectedOut []string
		expectdErr  error
	}{
		{
			builder: &mockContainer{
				runCommand: func() error { return testErr },
			},
			cfg:         testCfg,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1]},
			expectdErr:  testErr,
		},
		{
			builder: &mockContainer{
				runCommand: func() error { return nil },
				getId:      func() string { return "TESTID" },
			},
			cfg:         testCfg,
			debug:       true,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[3]},
			expectdErr:  nil,
		},
		{
			builder: &mockContainer{
				runCommand:  func() error { return nil },
				getId:       func() string { return "TESTID" },
				rmContainer: func() error { return testErr },
			},
			cfg:         testCfg,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[4]},
			expectdErr:  testErr,
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

		assert.Equal(t, tt.expectdErr, actualErr)
	}
}

func TestVerifyInputs(t *testing.T) {
	tests := []struct {
		cfg         *Config
		args        []string
		expectedErr error
	}{
		{
			cfg:         &Config{},
			args:        []string{},
			expectedErr: errors.ErrNotImplemented{},
		},
		{
			cfg:         &Config{},
			args:        []string{"."},
			expectedErr: errors.ErrWrongConfig{},
		},
		{
			cfg: &Config{
				Container: Container{
					Volume: "/tmp:/dst",
				},
			},
			args:        []string{"."},
			expectedErr: errors.ErrWrongConfig{},
		},
		{
			cfg: &Config{
				Container: Container{
					Volume: "/tmp",
				},
				Builder: Builder{
					UserDataFileName:      "user-data",
					NetworkConfigFileName: "net-conf",
				},
			},
			args:        []string{"."},
			expectedErr: nil,
		},
		{
			cfg: &Config{
				Container: Container{
					Volume: "/tmp:/dst:/dst1",
				},
				Builder: Builder{
					UserDataFileName:      "user-data",
					NetworkConfigFileName: "net-conf",
				},
			},
			args:        []string{"."},
			expectedErr: errors.ErrWrongConfig{},
		},
	}

	for _, tt := range tests {
		actualErr := verifyInputs(tt.cfg, tt.args)
		assert.Equal(t, tt.expectedErr, actualErr)
	}

}

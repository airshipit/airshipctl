package isogen

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/errors"
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
		"Creating cloud-init for ephemeral K8s\n",
		fmt.Sprintf("Running default container command. Mounted dir: [%s]\n", volBind),
		"ISO successfully built.\n",
		"Debug flag is set. Container TESTID stopped but not deleted.\n",
		"Removing container.\n",
	}

	tests := []struct {
		builder     *mockContainer
		cfg         *Config
		debug       bool
		expectedOut string
		expectdErr  error
	}{
		{
			builder: &mockContainer{
				runCommand: func() error { return testErr },
			},
			cfg:         testCfg,
			debug:       false,
			expectedOut: expOut[0] + expOut[1],
			expectdErr:  testErr,
		},
		{
			builder: &mockContainer{
				runCommand: func() error { return nil },
				getId:      func() string { return "TESTID" },
			},
			cfg:         testCfg,
			debug:       true,
			expectedOut: expOut[0] + expOut[1] + expOut[2] + expOut[3],
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
			expectedOut: expOut[0] + expOut[1] + expOut[2] + expOut[4],
			expectdErr:  testErr,
		},
	}

	for _, tt := range tests {
		actualOut := bytes.NewBufferString("")
		actualErr := generateBootstrapIso(bundle, tt.builder, tt.cfg, actualOut, tt.debug)

		errS := fmt.Sprintf("generateBootstrapIso should have return error %s, got %s", tt.expectdErr, actualErr)
		assert.Equal(t, actualErr, tt.expectdErr, errS)

		errS = fmt.Sprintf("generateBootstrapIso should have print %s, got %s", tt.expectedOut, actualOut)
		assert.Equal(t, actualOut.String(), tt.expectedOut, errS)
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
		actualErr := verifyInputs(tt.cfg, tt.args, bytes.NewBufferString(""))
		assert.Equal(t, tt.expectedErr, actualErr)
	}

}

package isogen

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
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
	testErr := fmt.Errorf("TestErr")
	expOut := []string{
		"Running default container command. Mounted dir: []\n",
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
			cfg:         &Config{},
			debug:       false,
			expectedOut: expOut[0],
			expectdErr:  testErr,
		},
		{
			builder: &mockContainer{
				runCommand: func() error { return nil },
				getId:      func() string { return "TESTID" },
			},
			cfg:         &Config{},
			debug:       true,
			expectedOut: expOut[0] + expOut[1] + expOut[2],
			expectdErr:  nil,
		},
		{
			builder: &mockContainer{
				runCommand:  func() error { return nil },
				getId:       func() string { return "TESTID" },
				rmContainer: func() error { return testErr },
			},
			cfg:         &Config{},
			debug:       false,
			expectedOut: expOut[0] + expOut[1] + expOut[3],
			expectdErr:  testErr,
		},
	}

	for _, tt := range tests {
		actualOut := bytes.NewBufferString("")
		actualErr := generateBootstrapIso(tt.builder, tt.cfg, actualOut, tt.debug)

		errS := fmt.Sprintf("generateBootstrapIso should have return error %s, got %s", tt.expectdErr, actualErr)
		assert.Equal(t, actualErr, tt.expectdErr, errS)

		errS = fmt.Sprintf("generateBootstrapIso should have print %s, got %s", tt.expectedOut, actualOut)
		assert.Equal(t, actualOut.String(), tt.expectedOut, errS)
	}
}

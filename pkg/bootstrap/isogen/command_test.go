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

package isogen_test

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
	"opendev.org/airship/airshipctl/pkg/bootstrap/isogen"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/testutil"
	testcontainer "opendev.org/airship/airshipctl/testutil/container"
	testdoc "opendev.org/airship/airshipctl/testutil/document"
)

const testID = "TESTID"

func TestBootstrapIso(t *testing.T) {
	bundle, err := document.NewBundleByPath("testdata/primary/site/test-site/ephemeral/bootstrap")
	require.NoError(t, err, "Building Bundle Failed")

	tempVol, cleanup := testutil.TempDir(t, "bootstrap-test")
	defer cleanup(t)

	volBind := tempVol + ":/dst"
	testErr := fmt.Errorf("TestErr")
	testCfg := &api.IsoConfiguration{
		IsoContainer: &api.IsoContainer{
			Volume:           volBind,
			ContainerRuntime: "docker",
		},
		Isogen: &api.Isogen{
			OutputFileName: "ephemeral.iso",
		},
	}
	testDoc := &testdoc.MockDocument{
		MockAsYAML: func() ([]byte, error) { return []byte("TESTDOC"), nil },
	}
	testIsogen := &testcontainer.MockContainer{
		MockRunCommand:  func() error { return nil },
		MockGetID:       func() string { return testID },
		MockRmContainer: func() error { return nil },
	}

	expOut := []string{
		"Creating cloud-init for ephemeral K8s",
		fmt.Sprintf("Running default container command. Mounted dir: [%s]", volBind),
		"ISO successfully built.",
		"Debug flag is set. Container TESTID stopped but not deleted.",
		"Removing container.",
	}

	tests := []struct {
		builder     *testcontainer.MockContainer
		cfg         *api.IsoConfiguration
		doc         *testdoc.MockDocument
		debug       bool
		expectedOut []string
		expectedErr error
	}{
		{
			builder: &testcontainer.MockContainer{
				MockRunCommand:        func() error { return testErr },
				MockWaitUntilFinished: func() error { return nil },
				MockRmContainer:       func() error { return nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1]},
			expectedErr: testErr,
		},
		{
			builder: &testcontainer.MockContainer{
				MockRunCommand:        func() error { return nil },
				MockGetID:             func() string { return "TESTID" },
				MockWaitUntilFinished: func() error { return nil },
				MockRmContainer:       func() error { return nil },
				MockGetContainerLogs:  func() (io.ReadCloser, error) { return ioutil.NopCloser(strings.NewReader("")), nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       true,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[3]},
			expectedErr: nil,
		},
		{
			builder: &testcontainer.MockContainer{
				MockRunCommand:        func() error { return nil },
				MockGetID:             func() string { return "TESTID" },
				MockRmContainer:       func() error { return testErr },
				MockWaitUntilFinished: func() error { return nil },
			},
			cfg:         testCfg,
			doc:         testDoc,
			debug:       false,
			expectedOut: []string{expOut[0], expOut[1], expOut[2], expOut[4]},
			expectedErr: testErr,
		},
		{
			builder: testIsogen,
			cfg:     testCfg,
			doc: &testdoc.MockDocument{
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
		bootstrapOpts := isogen.BootstrapIsoOptions{
			DocBundle: bundle,
			Builder:   tt.builder,
			Doc:       tt.doc,
			Cfg:       tt.cfg,
		}
		actualErr := bootstrapOpts.CreateBootstrapIso()
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
		cfg         *api.IsoConfiguration
		args        []string
		expectedErr error
	}{
		{
			name: "missing-container-field",
			cfg: &api.IsoConfiguration{
				IsoContainer: &api.IsoContainer{},
			},
			expectedErr: config.ErrMissingConfig{
				What: "Must specify volume bind for ISO builder container",
			},
		},
		{
			name: "invalid-host-path",
			cfg: &api.IsoConfiguration{
				IsoContainer: &api.IsoContainer{
					Volume: tempVol + ":/dst:/dst1",
				},
				Isogen: &api.Isogen{
					OutputFileName: "ephemeral.iso",
				},
			},
			expectedErr: config.ErrInvalidConfig{
				What: "Bad container volume format. Use hostPath:contPath",
			},
		},
		{
			name: "success",
			cfg: &api.IsoConfiguration{
				IsoContainer: &api.IsoContainer{
					Volume: tempVol,
				},
				Isogen: &api.Isogen{
					OutputFileName: "ephemeral.iso",
				},
			},
			expectedErr: nil,
		},
		{
			name: "success-using-output-file-default-",
			cfg: &api.IsoConfiguration{
				IsoContainer: &api.IsoContainer{
					Volume: tempVol,
				},
				Isogen: &api.Isogen{},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			actualErr := isogen.VerifyInputs(tt.cfg)
			assert.Equal(subTest, tt.expectedErr, actualErr)
		})
	}
}

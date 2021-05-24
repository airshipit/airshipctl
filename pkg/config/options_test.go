/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	testFormatError       = "wrong output format , must be one of yaml table"
	defaultCurrentContext = "a-context"
)

func TestContextOptionsValidate(t *testing.T) {
	tests := []struct {
		name        string
		testOptions config.ContextOptions
		expectError bool
	}{
		{
			name: "MissingName",
			testOptions: config.ContextOptions{
				Name: "",
			},
			expectError: true,
		},
		{
			name: "SettingCurrentContext",
			testOptions: config.ContextOptions{
				Name:           "testContext",
				CurrentContext: true,
			},
			expectError: false,
		},
		{
			name: "NoClusterType",
			testOptions: config.ContextOptions{
				Name: "testContext",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			err := tt.testOptions.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
func TestContextOptionsPrint(t *testing.T) {
	yamlOutput := `contexts:
  a-context:
    managementConfiguration: a-manageconf
    manifest: a-manifest
currentContext: a-context
`
	tests := []struct {
		name               string
		testContextOptions config.ContextOptions
		testConfig         config.Config
		expectedOutput     string
		expectedErr        string
	}{
		{
			name: "Wrong output format",
			testContextOptions: config.ContextOptions{
				Format: "",
			},
			testConfig:     config.Config{},
			expectedOutput: "",
			expectedErr:    testFormatError,
		},
		{
			name: "List contexts in table format",
			testContextOptions: config.ContextOptions{
				Name:           "",
				CurrentContext: false,
				Format:         "table",
			},
			testConfig: config.Config{
				CurrentContext: defaultCurrentContext,
				Contexts: map[string]*config.Context{
					"a-context": {Manifest: "a-manifest", ManagementConfiguration: "a-manageconf"},
					"b-context": {Manifest: "b-manifest", ManagementConfiguration: "b-manageconf"}},
			},
			expectedOutput: `CURRENT   NAME        MANIFEST     MANAGEMENTCONFIGURATION
*         a-context   a-manifest   a-manageconf
          b-context   b-manifest   b-manageconf
`,
		},
		{
			name: "List contexts in table format(Context name is given)",
			testContextOptions: config.ContextOptions{
				Name:           defaultCurrentContext,
				CurrentContext: false,
				Format:         "table",
			},
			testConfig: config.Config{
				CurrentContext: defaultCurrentContext,
				Contexts: map[string]*config.Context{
					"a-context": {Manifest: "a-manifest", ManagementConfiguration: "a-manageconf"},
					"b-context": {Manifest: "b-manifest", ManagementConfiguration: "b-manageconf"}},
			},
			expectedOutput: `CURRENT   NAME        MANIFEST     MANAGEMENTCONFIGURATION
*         a-context   a-manifest   a-manageconf
`,
		},
		{
			name: "List contexts in table format(CurrentContext is true)",
			testContextOptions: config.ContextOptions{
				Name:           "",
				CurrentContext: true,
				Format:         "table",
			},
			testConfig: config.Config{
				CurrentContext: defaultCurrentContext,
				Contexts: map[string]*config.Context{
					"a-context": {Manifest: "a-manifest", ManagementConfiguration: "a-manageconf"},
					"b-context": {Manifest: "b-manifest", ManagementConfiguration: "b-manageconf"}},
			},
			expectedOutput: `CURRENT   NAME        MANIFEST     MANAGEMENTCONFIGURATION
*         a-context   a-manifest   a-manageconf
`,
		},
		{
			name: "List contexts in table format(Wrong Name is given)",
			testContextOptions: config.ContextOptions{
				Name:           "wrong-context",
				CurrentContext: false,
				Format:         "table",
			},
			testConfig: config.Config{
				CurrentContext: defaultCurrentContext,
				Contexts: map[string]*config.Context{
					"a-context": {Manifest: "a-manifest", ManagementConfiguration: "a-manageconf"},
					"b-context": {Manifest: "b-manifest", ManagementConfiguration: "b-manageconf"}},
			},
			expectedOutput: `CURRENT   NAME   MANIFEST   MANAGEMENTCONFIGURATION
`,
			expectedErr: "context with name 'wrong-context'",
		},
		{
			name: "List contexts in yaml format",
			testContextOptions: config.ContextOptions{
				Name:           "",
				CurrentContext: false,
				Format:         "yaml",
			},
			testConfig: config.Config{
				CurrentContext: defaultCurrentContext,
				Contexts: map[string]*config.Context{
					"a-context": {Manifest: "a-manifest", ManagementConfiguration: "a-manageconf"}},
			},
			expectedOutput: yamlOutput,
		},
	}
	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := tt.testContextOptions.Print(&tt.testConfig, buf)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			out, err := ioutil.ReadAll(buf)
			fmt.Print(string(out))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOutput, string(out))
		})
	}
}

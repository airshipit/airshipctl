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

package util_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

func TestSetenv(t *testing.T) {
	tests := []struct {
		name        string
		input       []util.EnvVar
		expectedOut string
	}{
		{
			name: "success one env",
			input: []util.EnvVar{{
				Key:   "AIRSHIPCTL_TEST_KEY",
				Value: "AIRSHIPCTL_TEST_VALUE",
			}},
			expectedOut: "",
		},
		{
			name: "success multiple envs",
			input: []util.EnvVar{{
				Key:   "AIRSHIPCTL_TEST_KEY_1",
				Value: "AIRSHIPCTL_TEST_VALUE_1",
			}, {
				Key:   "AIRSHIPCTL_TEST_KEY_2",
				Value: "AIRSHIPCTL_TEST_VALUE_2"}},
			expectedOut: "",
		},
		{
			name: "fail to set",
			input: []util.EnvVar{{
				Key: "invalid_key\x00",
			}},
			expectedOut: "unable to set 'invalid_key\x00' env variable, reason 'setenv: invalid argument'",
		},
	}

	buf := &bytes.Buffer{}
	log.Init(false, buf)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			util.Setenv(tt.input...)
			require.Contains(t, buf.String(), tt.expectedOut)
			if tt.expectedOut == "" {
				for _, key := range tt.input {
					require.Equal(t, key.Value, os.Getenv(key.Key))
				}
			}
		})
	}
}

func TestUnsetenv(t *testing.T) {
	tests := []struct {
		name        string
		input       []util.EnvVar
		expectedOut string
	}{
		{
			name: "success one env",
			input: []util.EnvVar{{
				Key: "AIRSHIPCTL_TEST_KEY",
			}},
			expectedOut: "",
		},
		{
			name: "success multiple envs",
			input: []util.EnvVar{{
				Key: "AIRSHIPCTL_TEST_KEY_1",
			}, {
				Key: "AIRSHIPCTL_TEST_KEY_2"}},
			expectedOut: "",
		},
	}

	buf := &bytes.Buffer{}
	log.Init(false, buf)
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			util.Unsetenv(tt.input...)
			require.Contains(t, buf.String(), tt.expectedOut)
			if tt.expectedOut == "" {
				for _, key := range tt.input {
					require.Equal(t, "", os.Getenv(key.Key))
				}
			}
		})
	}
}

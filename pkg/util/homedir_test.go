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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/testutil"
)

func TestUserHomeDir(t *testing.T) {
	expectedTestDir, cleanup := testutil.TempDir(t, "test-home")
	defer cleanup(t)
	defer setHome(expectedTestDir)()

	assert.Equal(t, expectedTestDir, util.UserHomeDir())
}

// setHome sets the HOME environment variable to `path`, and returns a function
// that can be used to reset HOME to its original value
func setHome(path string) (resetHome func()) {
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", path)
	return func() {
		os.Setenv("HOME", oldHome)
	}
}

func TestExpandTilde(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedOut string
	}{
		{
			name:        "expand home path",
			input:       "~/.airship",
			expectedOut: filepath.Join(util.UserHomeDir(), "~/.airship"[1:]),
		},
		{
			name:        "expand just ~",
			input:       "~",
			expectedOut: util.UserHomeDir(),
		},
		{
			name:        "empty path",
			input:       "",
			expectedOut: "",
		},
		{
			name:        "absolute path",
			input:       "/.airship",
			expectedOut: "/.airship",
		},
		{
			name:        "tilde path",
			input:       "~.airship",
			expectedOut: "~.airship",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expectedOut, util.ExpandTilde(tt.input))
		})
	}
}

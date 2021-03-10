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
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

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
	t.Run("success home path", func(t *testing.T) {
		airshipDir := ".airship"
		dir := filepath.Join("~", airshipDir)
		expandedDir, err := util.ExpandTilde(dir)
		assert.NoError(t, err)
		homedir := path.Join(util.UserHomeDir(), airshipDir)
		assert.Equal(t, homedir, expandedDir)
	})

	t.Run("success nothing to expand", func(t *testing.T) {
		dir := "/home/ubuntu/.airship"
		expandedDir, err := util.ExpandTilde(dir)
		assert.NoError(t, err)
		assert.Equal(t, dir, expandedDir)
	})

	t.Run("error malformed path", func(t *testing.T) {
		malformedDir := "~.home/ubuntu/.airship"
		expandedDir, err := util.ExpandTilde(malformedDir)
		assert.Error(t, err)
		assert.Equal(t, "", expandedDir)
	})

	t.Run("success empty path", func(t *testing.T) {
		emptyDir := ""
		expandedDir, err := util.ExpandTilde(emptyDir)
		assert.NoError(t, err)
		assert.Equal(t, emptyDir, expandedDir)
	})
}

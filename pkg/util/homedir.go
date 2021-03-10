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

package util

import (
	"os"
	"path/filepath"
	"strings"
)

// UserHomeDir is a utility function that wraps os.UserHomeDir and returns no
// errors. If the user has no home directory, the returned value will be the
// empty string
func UserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return homeDir
}

// ExpandTilde expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~` or has no slash after tilde,
// the path is returned as-is.
func ExpandTilde(path string) string {
	// Just tilde - return current $HOME dir
	if path == "~" {
		return UserHomeDir()
	}
	// If path starts with ~/ - expand it
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(UserHomeDir(), path[1:])
	}

	// empty strings, absolute paths, ~<(dir/file)name> return as-is
	return path
}

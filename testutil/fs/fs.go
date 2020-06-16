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

package fs

import (
	"opendev.org/airship/airshipctl/pkg/document"
)

// MockFileSystem implements Filesystem
type MockFileSystem struct {
	MockRemoveAll func() error
	MockTempDir   func() (string, error)
	MockTempFile  func() (document.File, error)
	document.FileSystem
}

// RemoveAll Filesystem interface imlementation
func (fsys MockFileSystem) RemoveAll(string) error { return fsys.MockRemoveAll() }

// TempFile Filesystem interface imlementation
func (fsys MockFileSystem) TempFile(string, string) (document.File, error) {
	return fsys.MockTempFile()
}

// TempDir Filesystem interface imlementation
func (fsys MockFileSystem) TempDir(string, string) (string, error) {
	return fsys.MockTempDir()
}

// TestFile implements file
type TestFile struct {
	document.File
	MockName  func() string
	MockWrite func() (int, error)
	MockClose func() error
}

// Name File interface imlementation
func (f TestFile) Name() string { return f.MockName() }

// Write File interface imlementation
func (f TestFile) Write([]byte) (int, error) { return f.MockWrite() }

// Close File interface imlementation
func (f TestFile) Close() error { return f.MockClose() }

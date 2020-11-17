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

package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	fixtures "github.com/go-git/go-git-fixtures/v4"
	"github.com/stretchr/testify/require"
	kustfs "sigs.k8s.io/kustomize/api/filesys"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/fs"
)

// SetupTestFs help manufacture a fake file system for testing purposes. It
// will iterate over the files in fixtureDir, which is a directory relative
// to the tests themselves, and will write each of those files (preserving
// names) to an in-memory file system and return that fs
func SetupTestFs(t *testing.T, fixtureDir string) fs.FileSystem {
	t.Helper()

	x := &fs.Fs{FileSystem: kustfs.MakeFsInMemory()}

	files, err := ioutil.ReadDir(fixtureDir)
	require.NoErrorf(t, err, "Failed to read fixture directory %s", fixtureDir)
	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(fixtureDir, fileName)

		fileBytes, err := ioutil.ReadFile(filePath)
		require.NoErrorf(t, err, "Failed to read file %s, setting up testfs failed", filePath)
		err = x.WriteFile(filepath.Join("/", file.Name()), fileBytes)
		require.NoErrorf(t, err, "Failed to write file %s, setting up testfs failed", filePath)
	}
	return x
}

// NewTestBundle helps to create a new bundle with FakeFs containing documents from fixtureDir
func NewTestBundle(t *testing.T, fixtureDir string) document.Bundle {
	t.Helper()
	b, err := document.NewBundle(SetupTestFs(t, fixtureDir), "/")
	require.NoError(t, err, "Failed to build a bundle, setting up TestBundle failed")
	return b
}

// CleanUpGitFixtures removes any temp directories created by the go-git test fixtures
func CleanUpGitFixtures(t *testing.T) {
	if err := fixtures.Clean(); err != nil {
		t.Logf("Could not clean up git fixtures: %v", err)
	}
}

// TempDir creates a new temporary directory in the system's temporary file
// storage with a name beginning with prefix.
// It returns the path of the new directory and a function that can be used to
// easily clean up that directory
func TempDir(t *testing.T, prefix string) (path string, cleanup func(*testing.T)) {
	path, err := ioutil.TempDir("", prefix)
	require.NoError(t, err, "Failed to create a temporary directory")

	return path, func(tt *testing.T) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Logf("Could not clean up temp directory %q: %v", path, err)
		}
	}
}

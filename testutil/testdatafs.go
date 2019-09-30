package testutil

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

// SetupTestFs help manufacture a fake file system for testing purposes. It
// will iterate over the files in fixtureDir, which is a directory relative
// to the tests themselves, and will write each of those files (preserving
// names) to an in-memory file system and return that fs
func SetupTestFs(t *testing.T, fixtureDir string) fs.FileSystem {

	x := fs.MakeFakeFS()

	files, err := ioutil.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("Unable to read fixture directory %s: %v", fixtureDir, err)
	}
	for _, file := range files {
		fileName := file.Name()
		filePath := filepath.Join(fixtureDir, fileName)
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Error reading fixture %s: %v", filePath, err)
		}
		// nolint: errcheck
		err = x.WriteFile(filepath.Join("/", file.Name()), fileBytes)
		if err != nil {
			t.Fatalf("Error writing fixture %s: %v", filePath, err)
		}
	}
	return x

}

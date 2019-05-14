package util_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ian-howell/airshipctl/pkg/util"
)

func TestIsReadable(t *testing.T) {
	file, err := ioutil.TempFile(os.TempDir(), "airshipctl")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	defer os.Remove(file.Name())

	if err := util.IsReadable(file.Name()); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}

	expected := "permission denied"
	err = file.Chmod(0000)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if err := util.IsReadable(file.Name()); err == nil {
		t.Errorf("Expected '%s' error", expected)
	} else if err.Error() != "permission denied" {
		t.Errorf("Expected '%s' error, got '%s'", expected, err.Error())
	}

	expected = "no such file or directory"
	os.Remove(file.Name())
	if err := util.IsReadable(file.Name()); err == nil {
		t.Errorf("Expected '%s' error", expected)
	} else if err.Error() != expected {
		t.Errorf("Expected '%s' error, got '%s'", expected, err.Error())
	}
}

func TestReadDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "airshipctl-tests")
	if err != nil {
		t.Fatalf("Could not create a temporary directory: %s", err.Error())
	}
	defer os.RemoveAll(dir)

	testFiles := []string{
		"test1.txt",
		"test2.txt",
		"test3.txt",
	}

	for _, testFile := range testFiles {
		if err := ioutil.WriteFile(filepath.Join(dir, testFile), []byte("testdata"), 0666); err != nil {
			t.Fatalf("Could not create test file '%s': %s", testFile, err.Error())
		}
	}

	files, err := util.ReadDir(dir)
	if err != nil {
		t.Fatalf("Unexpected error while reading directory: %s", err.Error())
	}

	if len(files) != len(testFiles) {
		t.Errorf("Expected %d files, got %d", len(testFiles), len(files))
	}

	for _, testFile := range testFiles {
		found := false
		for _, actualFile := range files {
			if testFile == actualFile.Name() {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Could not find test file '%s'", testFile)
		}
	}

	os.RemoveAll(dir)

	if _, err := util.ReadDir(dir); err == nil {
		t.Error("Expected an error when reading non-existant directory")
	}
}

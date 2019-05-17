package util_test

import (
	"io/ioutil"
	"os"
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

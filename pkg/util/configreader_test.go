package util_test

import (
	"testing"

	"opendev.org/airship/airshipctl/pkg/util"
)

func TestReadYAMLFile(t *testing.T) {
	var actual map[string]interface{}
	if err := util.ReadYAMLFile("testdata/test.yaml", &actual); err != nil {
		t.Fatalf("Error while reading YAML: %s", err.Error())
	}
	expectedString := "test"
	actualString, ok := actual["testString"]
	if !ok {
		t.Fatalf("Missing \"testString\" attribute")
	}
	if actualString != expectedString {
		t.Errorf("Expected %s, got %s", expectedString, actualString)
	}
}

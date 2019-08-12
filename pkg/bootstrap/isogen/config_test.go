package isogen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToYaml(t *testing.T) {
	expectedBytes := []byte(`builder: {}
container:
  containerRuntime: docker
`)
	cnf := &Config{
		Container: Container{
			ContainerRuntime: "docker",
		},
	}

	actualBytes, _ := cnf.ToYAML()
	errS := fmt.Sprintf(
		"Call ToYAML should have returned %s, got %s",
		expectedBytes,
		actualBytes,
	)
	assert.Equal(t, actualBytes, expectedBytes, errS)
}

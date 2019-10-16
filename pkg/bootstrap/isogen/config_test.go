package isogen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	actualBytes, err := cnf.ToYAML()
	require.NoError(t, err)
	assert.Equal(t, actualBytes, expectedBytes)
}

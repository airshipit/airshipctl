package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/util"
)

func TestReadYAMLFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var actual map[string]interface{}
	err := util.ReadYAMLFile("testdata/test.yaml", &actual)
	require.NoError(err, "Error while reading YAML")

	actualString := actual["testString"]
	expectedString := "test"
	assert.Equal(expectedString, actualString)
}

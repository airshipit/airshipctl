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

	// test using an incorrect yaml
	err = util.ReadYAMLFile("testdata/incorrect.yaml", &actual)
	expectedString = "error converting YAML to JSON"
	require.Contains(err.Error(), expectedString)
}

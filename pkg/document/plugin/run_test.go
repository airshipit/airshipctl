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

package plugin_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/document/plugin"
)

func TestConfigureAndRun(t *testing.T) {
	testCases := []struct {
		pluginCfg     []byte
		expectedError string
		in            io.Reader
		out           io.Writer
	}{
		{
			pluginCfg:     []byte(""),
			expectedError: "plugin identified by /, Kind= was not found",
		},
		{
			pluginCfg: []byte(`---
apiVersion: airshipit.org/v1alpha1
kind: UnknownPlugin
spec:
  someField: someValue`),
			expectedError: "plugin identified by airshipit.org/v1alpha1, Kind=UnknownPlugin was not found",
		},
		{
			pluginCfg: []byte(`---
apiVersion: airshipit.org/v1alpha1
kind: BareMetalGenereator
spec: -
  someField: someValu`),
			expectedError: "error converting YAML to JSON: yaml: line 4: block sequence entries are not allowed in this context",
		},
	}

	for _, tc := range testCases {
		err := plugin.ConfigureAndRun(tc.pluginCfg, tc.in, tc.out)
		assert.EqualError(t, err, tc.expectedError)
	}
}

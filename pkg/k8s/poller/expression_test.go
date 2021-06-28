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

package poller_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/k8s/poller"
)

func TestNewExpression(t *testing.T) {
	testCases := []struct {
		name              string
		condition         string
		value             string
		obj               map[string]interface{}
		expectedResult    bool
		expectedErrString string
	}{
		{
			name:           "Success - value matched",
			condition:      "{.status}",
			obj:            map[string]interface{}{"status": "provisioned"},
			value:          "provisioned",
			expectedResult: true,
		},
		{
			name:           "Success - empty value",
			condition:      "{.status}",
			obj:            map[string]interface{}{"status": "provisioned"},
			expectedResult: true,
		},
		{
			name:              "Failed - invalid condition",
			condition:         "{*%.status}",
			expectedErrString: "unrecognized character in action",
		},
		{
			name:              "Failed - path not found in object",
			condition:         "{.status}",
			expectedErrString: "status is not found",
		},
	}
	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			exp := poller.Expression{
				Condition: tt.condition,
				Value:     tt.value,
			}

			res, err := exp.Match(tt.obj)
			assert.Equal(t, tt.expectedResult, res)
			if test.expectedErrString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectedErrString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

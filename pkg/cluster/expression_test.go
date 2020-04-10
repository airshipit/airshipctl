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

package cluster_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"opendev.org/airship/airshipctl/pkg/cluster"
)

func TestMatch(t *testing.T) {
	tests := map[string]struct {
		expression  cluster.Expression
		object      *unstructured.Unstructured
		expected    bool
		expectedErr error
	}{
		"healthy-object-matches-healthy": {
			expression: cluster.Expression{
				Condition: `@.status.health=="ok"`,
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "testversion/v1",
					"kind":       "TestObject",
					"metadata": map[string]interface{}{
						"name": "test-object",
					},
					"status": map[string]interface{}{
						"health": "ok",
					},
				},
			},
			expected: true,
		},
		"unhealthy-object-matches-unhealthy": {
			expression: cluster.Expression{
				Condition: `@.status.health=="ok"`,
			},
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "testversion/v1",
					"kind":       "TestObject",
					"metadata": map[string]interface{}{
						"name": "test-object",
					},
					"status": map[string]interface{}{
						"health": "not-ok",
					},
				},
			},
			expected: false,
		},
		"invalid-json-path-returns-error": {
			expression: cluster.Expression{
				Condition: `invalid JSON Path]`,
			},
			object: &unstructured.Unstructured{},
			expectedErr: cluster.ErrInvalidStatusCheck{
				What: `unable to parse jsonpath "invalid JSON Path]": ` +
					`unrecognized character in action: U+005D ']'`,
			},
		},
		"malformed-object-returns-error": {
			expression: cluster.Expression{
				Condition: `@.status.health=="ok"`,
			},
			object: &unstructured.Unstructured{},
			expectedErr: cluster.ErrInvalidStatusCheck{
				What: `failed to execute condition "@.status.health==\"ok\"" ` +
					`on object &{map[]}: status is not found`,
			},
		},
	}

	for testName, tt := range tests {
		tt := tt
		t.Run(testName, func(t *testing.T) {
			result, err := tt.expression.Match(tt.object)
			assert.Equal(t, tt.expectedErr, err)
			if tt.expectedErr == nil {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

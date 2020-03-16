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

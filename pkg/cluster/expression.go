package cluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

// An Expression is used to find information about a kubernetes resource. It
// evaluates to a boolean when matched against a resource.
type Expression struct {
	// A Condition describes a JSONPath filter which is matched against an
	// array containing a single resource.
	Condition string `json:"condition"`

	// jsonPath is used for the actual act of filtering on resources. It is
	// stored within the Expression as a means of memoization.
	jsonPath *jsonpath.JSONPath
}

// Match returns true if the given object matches the parsed jsonpath object.
// An error is returned if the Expression's condition is not a valid JSONPath
// as defined here: https://goessner.net/articles/JsonPath.
func (e *Expression) Match(obj runtime.Unstructured) (bool, error) {
	// NOTE(howell): JSONPath filters only work on lists. This means that
	// in order to check if a certain condition is met for obj, we need to
	// put obj into an list, then see if the filter catches obj.
	const listName = "items"

	// Parse lazily
	if e.jsonPath == nil {
		jp := jsonpath.New("status-check")

		// The condition must be a filter on a list
		itemAsArray := fmt.Sprintf("{$.%s[?(%s)]}", listName, e.Condition)
		err := jp.Parse(itemAsArray)
		if err != nil {
			return false, ErrInvalidStatusCheck{
				What: fmt.Sprintf("unable to parse jsonpath %q: %v", e.Condition, err.Error()),
			}
		}
		e.jsonPath = jp
	}

	// Filters only work on lists
	list := map[string]interface{}{
		listName: []interface{}{obj.UnstructuredContent()},
	}
	results, err := e.jsonPath.FindResults(list)
	if err != nil {
		return false, ErrInvalidStatusCheck{
			What: fmt.Sprintf("failed to execute condition %q on object %v: %v", e.Condition, obj, err),
		}
	}
	return len(results[0]) == 1, nil
}

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

package poller

import (
	"fmt"

	"k8s.io/client-go/util/jsonpath"
)

// An Expression is used to find information about a kubernetes resource. It
// evaluates to a boolean when matched against a resource.
type Expression struct {
	// A Condition describes a JSONPath filter which is matched against an
	// array containing a single resource.
	Condition string
	Value     string

	// jsonPath is used for the actual act of filtering on resources. It is
	// stored within the Expression as a means of memoization.
	jsonPath *jsonpath.JSONPath
}

// Match returns true if the given object matches the parsed jsonpath object.
// An error is returned if the Expression's condition is not a valid JSONPath
// as defined here: https://goessner.net/articles/JsonPath.
func (e *Expression) Match(obj map[string]interface{}) (bool, error) {
	// Parse lazily
	if e.jsonPath == nil {
		jp := jsonpath.New("status-check")

		err := jp.Parse(e.Condition)
		if err != nil {
			return false, err
		}
		e.jsonPath = jp
	}

	results, err := e.jsonPath.FindResults(obj)
	if err != nil {
		return false, err
	}

	if e.Value != "" {
		return len(results[0]) == 1 && fmt.Sprintf("%s", results[0][0].Interface()) == e.Value, nil
	}

	return len(results[0]) == 1, nil
}

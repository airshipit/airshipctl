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

package kyamlutils_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"opendev.org/airship/airshipctl/pkg/document/plugin/kyamlutils"
)

func TestGetterFilter(t *testing.T) {
	obj, err := yaml.Parse(`apiVersion: apps/v1
kind: SomeKind
metadata:
  name: resource-name
spec:
  path:
    to:
      some:
        scalar:
          value: 300
        nonScalar:
          key: val
        array:
          - i1
          - i2
keyWith.dot:
  - subkeyWith.dot: val1
    name: obj1
  - subkeyWith.dot: val2
    name: obj2
listOfStrings:
  - string1
  - string2
  - string3
listOfObjects:
  - name: obj1
    value: 10
  - name: obj2
    value: 20
  - name: obj3
    value: 30
listOfComplexObjects:
  - name: o1
    value:
      someKey1: valO11
      someKey2: valO12
    value1:
      someKey1: valO11
      someKey2: valO12
  - name: o2
    value:
      someKey0: false
      someKey1: valO21
      someKey2: valO22
sameObjects:
  - name: otherName
    val: otherValue
  - name: o1
    val: val1
  - name: o1
    val: val1`)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		query       string
		expectedErr string
		expected    string
	}{
		// Parser cases
		{
			name:        "Multiple expressions",
			query:       "{.a.b} {.x.y}",
			expectedErr: "query must contain one expression",
		},
		// Field cases
		{
			name:     "Path to scalar value",
			query:    "{.spec.path.to.some.scalar.value}",
			expected: "300",
		},
		{
			name:  "Key name with dot",
			query: "{.keyWith\\.dot}",
			expected: `
- subkeyWith.dot: val1
  name: obj1
- subkeyWith.dot: val2
  name: obj2`[1:],
		},
		{
			name:  "Field from array of scalar values",
			query: "{.listOfStrings.value}",
			expectedErr: "wrong Node Kind for listOfStrings " +
				"expected: MappingNode was SequenceNode: value: {- string1\n- string2\n- string3}",
		},
		{
			name:  "Non Existent path",
			query: "{.a.b}",
		},
		// Array cases
		{
			name:  "Array element negative",
			query: "{.listOfObjects[-1]}",
			expected: `
name: obj3
value: 30`[1:],
		},
		{
			name:  "Array element",
			query: "{.listOfObjects[0]}",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:     "Array element value",
			query:    "{.listOfObjects[0].value}",
			expected: "10",
		},
		{
			name:  "From element to end",
			query: "{.listOfObjects[1:]}",
			expected: `
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Segment",
			query: "{.listOfObjects[1:3]}",
			expected: `
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Segment with negative right",
			query: "{.listOfObjects[1:-1]}",
			expected: `
name: obj2
value: 20`[1:],
		},
		{
			name:  "From start to element",
			query: "{.listOfObjects[:2]}",
			expected: `
- name: obj1
  value: 10
- name: obj2
  value: 20`[1:],
		},
		{
			name:  "Iterate with step",
			query: "{.listOfObjects[::2]}",
			expected: `
- name: obj1
  value: 10
- name: obj3
  value: 30`[1:],
		},
		{
			name:     "Empty list",
			query:    "{.listOfObjects[1:1]}",
			expected: "",
		},
		{
			name:        "Access to non array",
			query:       "{.metadata[1]}",
			expectedErr: "wrong Node Kind for metadata expected: SequenceNode was MappingNode: value: {name: resource-name}",
		},
		{
			name:        "Out of left bound",
			query:       "{.listOfObjects[100]}",
			expectedErr: "array index out of bounds: index 100, length 3",
		},
		{
			name:        "Out of right bound",
			query:       "{.listOfObjects[:100]}",
			expectedErr: "array index out of bounds: index 99, length 3",
		},
		{
			name:        "Bad bounds",
			query:       "{.listOfObjects[2:1]}",
			expectedErr: "starting index 2 is greater than ending index 1",
		},
		{
			name:        "Negative step",
			query:       "{.listOfObjects[::-2]}",
			expectedErr: "step must be > 0",
		},
		// Filter cases
		{
			name:  "Filter by key",
			query: "{.listOfObjects[?(.name == 'obj1')]}",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:  "Filter by key name with dot",
			query: "{.keyWith\\.dot[?(.subkeyWith\\.dot == 'val2')]}",
			expected: `
subkeyWith.dot: val2
name: obj2`[1:],
		},
		{
			name:  "Filter eq int",
			query: "{.listOfObjects[?(.value == 10)]}",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:     "Filter by content",
			query:    "{.listOfStrings[?(. == 'string1')]}",
			expected: "string1",
		},
		{
			name:        "Query parse error",
			query:       "{.listOfStrings[?(. == 'string1'",
			expectedErr: "unterminated filter",
		},
		{
			name:        "Filter non-array",
			query:       "{.metadata[?(.name == 'name')]}",
			expectedErr: "wrong Node Kind for metadata expected: SequenceNode was MappingNode: value: {name: resource-name}",
		},
		{
			name:        "Filter condition non-array",
			query:       "{.listOfObjects[?(.[0] == 'name')]}",
			expectedErr: "wrong Node Kind for  expected: SequenceNode was MappingNode: value: {name: obj1\nvalue: 10}",
		},
		{
			name:     "Filter eq non-scalars",
			query:    "{.listOfComplexObjects[?(.value == 'name')]}",
			expected: "",
		},
		{
			name:  "Filter compare non-scalars",
			query: "{.listOfComplexObjects[?(.value > 'name')]}",
			expectedErr: "wrong Node Kind for value expected: ScalarNode was MappingNode: value: {someKey1: valO11\n" +
				"someKey2: valO12}",
		},
		{
			name:        "Filter unknown operator",
			query:       "{.listOfStrings[?(. >> 'name')]}",
			expectedErr: "unrecognized filter operator >>",
		},
		{
			name:  "Filter gt int",
			query: "{.listOfObjects[?(.value > 10)]}",
			expected: `
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Filter gt float",
			query: "{.listOfObjects[?(.value > 7.5)]}",
			expected: `
- name: obj1
  value: 10
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Filter ge float",
			query: "{.listOfObjects[?(.value >= 20.0)]}",
			expected: `
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Filter lt int",
			query: "{.listOfObjects[?(20 < .value)]}",
			expected: `
name: obj3
value: 30`[1:],
		},
		{
			name:  "Filter eq objects",
			query: "{.listOfObjects[?(@ == @)]}",
			expected: `
- name: obj1
  value: 10
- name: obj2
  value: 20
- name: obj3
  value: 30`[1:],
		},
		{
			name:  "Filter eq sub-objects",
			query: "{.listOfComplexObjects[?(.value == .value1)]}",
			expected: `
name: o1
value:
  someKey1: valO11
  someKey2: valO12
value1:
  someKey1: valO11
  someKey2: valO12`[1:],
		},
		{
			name:        "Filter not eq error",
			query:       "{.listOfObjects[?(.name != 10)]}",
			expectedErr: `strconv.Atoi: parsing "obj1": invalid syntax`,
		},
		{
			name:  "Filter eq bool",
			query: "{.listOfComplexObjects[?(.value.someKey0 == false)]}",
			expected: `
name: o2
value:
  someKey0: false
  someKey1: valO21
  someKey2: valO22`[1:],
		},
		// Wildcard cases
		{
			name:  "Get map by wildcard",
			query: "{.spec.path.to.some.*}",
			expected: `
- value: 300
- key: val
- - i1
  - i2`[1:],
		},
		{
			name:  "Wildcard list",
			query: "{.listOfStrings.*}",
			expected: `
- string1
- string2
- string3`[1:],
		},
		// v1 queries cases
		{
			name:  "Array element v1",
			query: "listOfObjects[0]",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:     "Array element value v1",
			query:    "listOfObjects[0].value",
			expected: "10",
		},
		{
			name:  "Array element by numeric key v1",
			query: "listOfObjects.0",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:     "Array element value by numeric key v1",
			query:    "listOfObjects.0.value",
			expected: "10",
		},
		{
			name:  "Array filter v1",
			query: "listOfObjects[name=obj1]",
			expected: `
name: obj1
value: 10`[1:],
		},
		{
			name:     "Array filter get all value v1",
			query:    "listOfObjects[name=obj1].value",
			expected: "10",
		},
		{
			name:  "Array filter get same value v1",
			query: "sameObjects[name=o1].val",
			expected: `
- val1
- val1`[1:],
		},
		{
			name:  "Array filter get first value v1",
			query: "['keyWith.dot'][subkeyWith.dot=val1]",
			expected: `
subkeyWith.dot: val1
name: obj1`[1:],
		},
		{
			name:        "Nested parentheses",
			query:       "spec[path[0]]",
			expectedErr: "failed to convert v1 path 'spec[path[0]]' to jsonpath. nested parentheses are not allowed",
		},
		{
			name:        "Invalid closing parentheses",
			query:       "spec[path]]",
			expectedErr: "failed to convert v1 path 'spec[path]]' to jsonpath. invalid field path",
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			getter := kyamlutils.JSONPathFilter{Path: tt.query}
			actualObj, err := obj.Pipe(getter)
			errS := ""
			if err != nil {
				errS = err.Error()
			}
			assert.Equal(t, tc.expectedErr, errS)
			actualString, err := actualObj.String()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, strings.TrimSuffix(actualString, "\n"))
		})
	}
}

func TestFilterCreate(t *testing.T) {
	data := `apiVersion: apps/v1
kind: SomeKind
metadata:
  name: resource-name`
	testCases := []struct {
		name   string
		query  string
		create bool
	}{
		{
			name:   "Create if not exists",
			query:  "{.spec.not.exists}",
			create: true,
		},
		{
			name:  "Not exists",
			query: "{.spec.not.exists}",
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			obj, err := yaml.Parse(data)
			require.NoError(t, err)
			getter := kyamlutils.JSONPathFilter{Path: tt.query, Create: tt.create}
			actualObj, err := obj.Pipe(getter)
			require.NoError(t, err)
			if tt.create {
				assert.NotNil(t, actualObj)
			} else {
				assert.Nil(t, actualObj)
			}
		})
	}
}

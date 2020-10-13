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

package kyamlutils

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// ErrIndexOutOfBound returned if JSON path points to a wrong index of a list
type ErrIndexOutOfBound struct {
	Index  int
	Length int
}

func (e ErrIndexOutOfBound) Error() string {
	return fmt.Sprintf("array index out of bounds: index %d, length %d", e.Index, e.Length)
}

// ErrBadQueryFormat raised for JSON query errors
type ErrBadQueryFormat struct {
	Msg string
}

func (e ErrBadQueryFormat) Error() string {
	return e.Msg
}

// ErrLookup raised if error occurs during walk through yaml resource
type ErrLookup struct {
	Msg string
}

func (e ErrLookup) Error() string {
	return e.Msg
}

// ErrNotScalar returned if value defined by key in JSON path is not scalar
// Error can be returned for filter queries
type ErrNotScalar struct {
	Node *yaml.Node
}

func (e ErrNotScalar) Error() string {
	return fmt.Sprintf("%#v is not scalar", e.Node)
}

// ErrQueryConversion returned by query conversion function
type ErrQueryConversion struct {
	Msg   string
	Query string
}

func (e ErrQueryConversion) Error() string {
	return fmt.Sprintf("failed to convert v1 path '%s' to jsonpath. %s", e.Query, e.Msg)
}

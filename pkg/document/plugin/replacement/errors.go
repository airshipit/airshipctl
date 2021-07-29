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

package replacement

import (
	"fmt"
	"reflect"
	"strings"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
)

// ErrBadConfiguration returned in case of plugin misconfiguration
type ErrBadConfiguration struct {
	Msg string
}

func (e ErrBadConfiguration) Error() string {
	return e.Msg
}

// ErrMultipleResources returned if multiple resources were found
type ErrMultipleResources struct {
	ObjRef *airshipv1.Target
}

func (e ErrMultipleResources) Error() string {
	return fmt.Sprintf("found more than one resources matching identified by %s", printFields(e.ObjRef))
}

// ErrSourceNotFound returned if a replacement source resource does not exist in resource map
type ErrSourceNotFound struct {
	ObjRef *airshipv1.Target
}

func (e ErrSourceNotFound) Error() string {
	return fmt.Sprintf("failed to find any source resources identified by %s", printFields(e.ObjRef))
}

// ErrTargetNotFound returned if a replacement target resource does not exist in the resource map
type ErrTargetNotFound struct {
	ObjRef *airshipv1.Selector
}

func (e ErrTargetNotFound) Error() string {
	return fmt.Sprintf("failed to find any target resources identified by %s", printFields(e.ObjRef))
}

// ErrPatternSubstring returned in case of issues with sub-string pattern substitution
type ErrPatternSubstring struct {
	Msg string
}

func (e ErrPatternSubstring) Error() string {
	return e.Msg
}

// ErrBase64Decoding returned incorrect base64 encoded string received for `kind: Secret`
type ErrBase64Decoding struct {
	Msg string
}

func (e ErrBase64Decoding) Error() string {
	return e.Msg
}

// ErrIndexOutOfBound returned if JSON path points to a wrong index of a list
type ErrIndexOutOfBound struct {
	Index  int
	Length int
}

func (e ErrIndexOutOfBound) Error() string {
	return fmt.Sprintf("array index out of bounds: index %d, length %d", e.Index, e.Length)
}

func printFields(objRef interface{}) string {
	val := reflect.ValueOf(objRef).Elem()
	valType := val.Type()
	var res []string
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.String() != "" {
			res = append(res, fmt.Sprintf("%s: %v", valType.Field(i).Name, field.Interface()))
		}
	}
	return strings.Join(res, " ")
}

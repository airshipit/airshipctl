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
	"strings"

	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
)

// ErrTypeMismatch is returned for type conversion errors. This error
// is raised if JSON path element points to a wrong data structure e.g.
// JSON path 'a.b[x=y]c' considers that there is a list of maps under key 'b'
// therefore ErrTypeMismatch will be returned for following structure
//
// a:
//   b:
//     - 'some string'
type ErrTypeMismatch struct {
	Actual      interface{}
	Expectation string
}

func (e ErrTypeMismatch) Error() string {
	return fmt.Sprintf("%#v %s", e.Actual, e.Expectation)
}

// ErrBadConfiguration returned in case of plugin misconfiguration
type ErrBadConfiguration struct {
	Msg string
}

func (e ErrBadConfiguration) Error() string {
	return e.Msg
}

// ErrMultipleResources returned if multiple resources were found
type ErrMultipleResources struct {
	ResList []*resource.Resource
}

func (e ErrMultipleResources) Error() string {
	return fmt.Sprintf("found more than one resources matching from %v", e.ResList)
}

// ErrSourceNotFound returned if a replacement source resource does not exist in resource map
type ErrSourceNotFound struct {
	ObjRef *types.Target
}

func (e ErrSourceNotFound) Error() string {
	keys := [5]string{"Group:", "Version:", "Kind:", "Name:", "Namespace:"}
	values := [5]string{e.ObjRef.Group, e.ObjRef.Version, e.ObjRef.Kind, e.ObjRef.Name, e.ObjRef.Namespace}

	var resFilter string
	for i, key := range keys {
		if values[i] != "" {
			resFilter += key + values[i] + " "
		}
	}
	return fmt.Sprintf("failed to find any source resources identified by %s", strings.TrimSpace(resFilter))
}

// ErrTargetNotFound returned if a replacement target resource does not exist in the resource map
type ErrTargetNotFound struct {
	ObjRef *types.Selector
}

func (e ErrTargetNotFound) Error() string {
	keys := [7]string{"Group:", "Version:", "Kind:", "Name:", "Namespace:",
		"AnnotationSelector:", "LabelSelector:"}
	values := [7]string{e.ObjRef.Group, e.ObjRef.Version, e.ObjRef.Kind, e.ObjRef.Name,
		e.ObjRef.Namespace, e.ObjRef.AnnotationSelector, e.ObjRef.LabelSelector}

	var resFilter string
	for i, key := range keys {
		if values[i] != "" {
			resFilter += key + values[i] + " "
		}
	}
	return fmt.Sprintf("failed to find any target resources identified by %s", strings.TrimSpace(resFilter))
}

// ErrPatternSubstring returned in case of issues with sub-string pattern substitution
type ErrPatternSubstring struct {
	Msg string
}

func (e ErrPatternSubstring) Error() string {
	return e.Msg
}

// ErrIndexOutOfBound returned if JSON path points to a wrong index of a list
type ErrIndexOutOfBound struct {
	Index int
}

func (e ErrIndexOutOfBound) Error() string {
	return fmt.Sprintf("index %v is out of bound", e.Index)
}

// ErrMapNotFound returned if map specified in fieldRef option was not found in a list
type ErrMapNotFound struct {
	Key, Value, ListKey string
}

func (e ErrMapNotFound) Error() string {
	return fmt.Sprintf("unable to find map key '%s' with the value '%s' in list under '%s' key",
		e.Key, e.Value, e.ListKey)
}

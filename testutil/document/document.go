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

package document

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// MockDocument implements Document for unit test purposes
type MockDocument struct {
	MockAnnotate       func()
	MockAsYAML         func() ([]byte, error)
	MockGetAnnotations func() map[string]string
	MockGetBool        func() (bool, error)
	MockGetFloat64     func() (float64, error)
	MockGetGroup       func() string
	MockGetInt64       func() (int64, error)
	MockGetKind        func() string
	MockGetLabels      func() map[string]string
	MockGetMap         func() (map[string]interface{}, error)
	MockGetName        func() string
	MockGetNamespace   func() string
	MockGetSlice       func() ([]interface{}, error)
	MockGetString      func() (string, error)
	MockGetStringMap   func() (map[string]string, error)
	MockGetStringSlice func() ([]string, error)
	MockGetVersion     func() string
	MockLabel          func()
	MockMarshalJSON    func() ([]byte, error)
	MockToObject       func() error
	MockToAPIObject    func() error
	MockGetFieldValue  func() (interface{}, error)
}

// Annotate Document interface implementation for unit test purposes
func (md *MockDocument) Annotate(_ map[string]string) {
	md.MockAnnotate()
}

// AsYAML Document interface implementation for unit test purposes
func (md *MockDocument) AsYAML() ([]byte, error) {
	return md.MockAsYAML()
}

// GetAnnotations Document interface implementation for unit test purposes
func (md *MockDocument) GetAnnotations() map[string]string {
	return md.MockGetAnnotations()
}

// GetBool Document interface implementation for unit test purposes
func (md *MockDocument) GetBool(_ string) (bool, error) {
	return md.MockGetBool()
}

// GetFloat64 Document interface implementation for unit test purposes
func (md *MockDocument) GetFloat64(_ string) (float64, error) {
	return md.MockGetFloat64()
}

// GetFieldValue Document interface implementation for unit test purposes
func (md *MockDocument) GetFieldValue(_ string) (interface{}, error) {
	return md.MockGetFieldValue()
}

// GetGroup Document interface implementation for unit test purposes
func (md *MockDocument) GetGroup() string {
	return md.MockGetGroup()
}

// GetInt64 Document interface implementation for unit test purposes
func (md *MockDocument) GetInt64(_ string) (int64, error) {
	return md.MockGetInt64()
}

// GetKind Document interface implementation for unit test purposes
func (md *MockDocument) GetKind() string {
	return md.MockGetKind()
}

// GetLabels Document interface implementation for unit test purposes
func (md *MockDocument) GetLabels() map[string]string {
	return md.MockGetLabels()
}

// GetMap Document interface implementation for unit test purposes
func (md *MockDocument) GetMap(_ string) (map[string]interface{}, error) {
	return md.MockGetMap()
}

// GetName Document interface implementation for unit test purposes
func (md *MockDocument) GetName() string {
	return md.MockGetName()
}

// GetNamespace Document interface implementation for unit test purposes
func (md *MockDocument) GetNamespace() string {
	return md.MockGetNamespace()
}

// GetSlice Document interface implementation for unit test purposes
func (md *MockDocument) GetSlice(_ string) ([]interface{}, error) {
	return md.MockGetSlice()
}

// GetString Document interface implementation for unit test purposes
func (md *MockDocument) GetString(_ string) (string, error) {
	return md.MockGetString()
}

// GetStringMap Document interface implementation for unit test purposes
func (md *MockDocument) GetStringMap(_ string) (map[string]string, error) {
	return md.MockGetStringMap()
}

// GetStringSlice Document interface implementation for unit test purposes
func (md *MockDocument) GetStringSlice(_ string) ([]string, error) {
	return md.MockGetStringSlice()
}

// GetVersion Document interface implementation for unit test purposes
func (md *MockDocument) GetVersion() string {
	return md.MockGetVersion()
}

// Label Document interface implementation for unit test purposes
func (md *MockDocument) Label(_ map[string]string) {
	md.MockLabel()
}

// MarshalJSON Document interface implementation for unit test purposes
func (md *MockDocument) MarshalJSON() ([]byte, error) {
	return md.MockMarshalJSON()
}

// ToObject Document interface implementation for unit test purposes
func (md *MockDocument) ToObject(_ interface{}) error {
	return md.MockToObject()
}

// ToAPIObject Document interface implementation for unit test purposes
func (md *MockDocument) ToAPIObject(obj runtime.Object, scheme *runtime.Scheme) error {
	return md.MockToAPIObject()
}

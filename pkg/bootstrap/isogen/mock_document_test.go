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

package isogen

import (
	"k8s.io/apimachinery/pkg/runtime"
)

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
}

func (md *MockDocument) Annotate(_ map[string]string) {
	md.MockAnnotate()
}

func (md *MockDocument) AsYAML() ([]byte, error) {
	return md.MockAsYAML()
}

func (md *MockDocument) GetAnnotations() map[string]string {
	return md.MockGetAnnotations()
}

func (md *MockDocument) GetBool(_ string) (bool, error) {
	return md.MockGetBool()
}

func (md *MockDocument) GetFloat64(_ string) (float64, error) {
	return md.MockGetFloat64()
}

func (md *MockDocument) GetGroup() string {
	return md.MockGetGroup()
}

func (md *MockDocument) GetInt64(_ string) (int64, error) {
	return md.MockGetInt64()
}

func (md *MockDocument) GetKind() string {
	return md.MockGetKind()
}

func (md *MockDocument) GetLabels() map[string]string {
	return md.MockGetLabels()
}

func (md *MockDocument) GetMap(_ string) (map[string]interface{}, error) {
	return md.MockGetMap()
}

func (md *MockDocument) GetName() string {
	return md.MockGetName()
}

func (md *MockDocument) GetNamespace() string {
	return md.MockGetNamespace()
}

func (md *MockDocument) GetSlice(_ string) ([]interface{}, error) {
	return md.MockGetSlice()
}

func (md *MockDocument) GetString(_ string) (string, error) {
	return md.MockGetString()
}

func (md *MockDocument) GetStringMap(_ string) (map[string]string, error) {
	return md.MockGetStringMap()
}

func (md *MockDocument) GetStringSlice(_ string) ([]string, error) {
	return md.MockGetStringSlice()
}

func (md *MockDocument) GetVersion() string {
	return md.MockGetVersion()
}

func (md *MockDocument) Label(_ map[string]string) {
	md.MockLabel()
}

func (md *MockDocument) MarshalJSON() ([]byte, error) {
	return md.MockMarshalJSON()
}

func (md *MockDocument) ToObject(_ interface{}) error {
	return md.MockToObject()
}

func (md *MockDocument) ToAPIObject(obj runtime.Object, scheme *runtime.Scheme) error {
	return md.MockToAPIObject()
}

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

package implementations

import (
	"fmt"
)

// ErrVersionNotDefined is returned when requested version is not present in repository
type ErrVersionNotDefined struct {
	Version string
}

func (e ErrVersionNotDefined) Error() string {
	return fmt.Sprintf(`version %s is not defined in the repository`, e.Version)
}

// ErrNoVersionsAvailable is returned when version map is empty or not defined
type ErrNoVersionsAvailable struct {
	Versions map[string]string
}

func (e ErrNoVersionsAvailable) Error() string {
	return fmt.Sprintf(`version map is empty or not defined, %v`, e.Versions)
}

// ErrValueForVariableNotSet is returned when version map is empty or not defined
type ErrValueForVariableNotSet struct {
	Variable string
}

func (e ErrValueForVariableNotSet) Error() string {
	return fmt.Sprintf("value for variable %q is not set", e.Variable)
}

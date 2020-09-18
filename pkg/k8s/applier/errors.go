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

package applier

import (
	"fmt"
)

// ErrApply returned for not implemented features
type ErrApply struct {
	errors []error
}

func (e ErrApply) Error() string {
	// TODO make printing more readable here
	return fmt.Sprintf("Applying of resources to kubernetes cluster has failed, errors are:\n%v", e.errors)
}

// ErrNilBundle returned when bundle is nil
type ErrNilBundle struct {
}

func (e ErrNilBundle) Error() string {
	return "nil bundle provided"
}

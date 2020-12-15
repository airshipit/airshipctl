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

package executors

import (
	"fmt"
)

// ErrUnknownExecutorAction is returned for unknown action parameter
// in clusterctl configuration document
type ErrUnknownExecutorAction struct {
	Action string
}

func (e ErrUnknownExecutorAction) Error() string {
	return fmt.Sprintf("unknown action type '%s'", e.Action)
}

// ErrIsoGenNilBundle is returned when isogen executor is not provided with bundle
type ErrIsoGenNilBundle struct {
}

func (e ErrIsoGenNilBundle) Error() string {
	return "Cannot build iso with empty bundle, no data source is available"
}

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

package checkexpiration

import "fmt"

// ErrInvalidFormat is called when the user provides format other than yaml/json
type ErrInvalidFormat struct {
	RequestedFormat string
}

func (e ErrInvalidFormat) Error() string {
	return fmt.Sprintf("invalid output format specified %s. Allowed values are json|yaml", e.RequestedFormat)
}

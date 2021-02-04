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

package phase

import "fmt"

// ErrRenderTooManyArgs returned when more than 1 argument is provided as argument to render command
type ErrRenderTooManyArgs struct {
	Count int
}

func (e *ErrRenderTooManyArgs) Error() string {
	return fmt.Sprintf("accepts 1 or less arg(s), received %d", e.Count)
}

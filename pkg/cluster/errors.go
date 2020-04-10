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

package cluster

import "fmt"

// ErrInvalidStatusCheck denotes that something went wrong while handling a
// status-check annotation.
type ErrInvalidStatusCheck struct {
	What string
}

func (err ErrInvalidStatusCheck) Error() string {
	return fmt.Sprintf("invalid status-check: %s", err.What)
}

// ErrResourceNotFound is used when a resource is requrested from a StatusMap,
// but that resource can't be found
type ErrResourceNotFound struct {
	Resource string
}

func (err ErrResourceNotFound) Error() string {
	return fmt.Sprintf("could not find a status for resource %q", err.Resource)
}

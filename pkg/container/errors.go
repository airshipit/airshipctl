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

package container

import (
	"fmt"
)

// ErrEmptyImageList returned if no image defined in filter found
type ErrEmptyImageList struct {
}

func (e ErrEmptyImageList) Error() string {
	return "Image List Error. No images filtered"
}

// ErrRunContainerCommand returned if container command
// exited with non-zero code
type ErrRunContainerCommand struct {
	Cmd string
}

func (e ErrRunContainerCommand) Error() string {
	return fmt.Sprintf("Command error. run '%s' for details", e.Cmd)
}

// ErrContainerDrvNotSupported returned if desired CRI is not supported
type ErrContainerDrvNotSupported struct {
	Driver string
}

func (e ErrContainerDrvNotSupported) Error() string {
	return fmt.Sprintf("Driver %s is not supported", e.Driver)
}

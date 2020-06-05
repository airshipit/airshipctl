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

package ifc

import (
	"io"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Executor interface should be implemented by each runner
type Executor interface {
	Run(dryrun, debug bool) error
	Render(io.Writer) error
	Validate() error
	Wait() error
}

// ExecutorFactory for executor instantiation
// First argument is document object which represents executor
// configuration.
// Second argument is document bundle used by executor.
// Third argument airship configuration settings since each phase
// has to be aware of execution context and global settings
type ExecutorFactory func(
	document.Document,
	document.Bundle,
	*environment.AirshipCTLSettings,
) (Executor, error)

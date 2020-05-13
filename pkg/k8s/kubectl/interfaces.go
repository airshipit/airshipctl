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

package kubectl

import (
	"opendev.org/airship/airshipctl/pkg/document"
)

//TODO(sy495p): type name "Interface" is too generic. Use a more specific name

// Interface provides a abstraction layer built on top of kubectl libraries
// to implement kubectl subcommands as kubectl apply
type Interface interface {
	Apply(docs []document.Document, ao *ApplyOptions) error
	ApplyOptions() (*ApplyOptions, error)
}

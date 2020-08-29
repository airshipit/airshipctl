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

package render

// Settings for document rendering
type Settings struct {
	// Label filters documents by label string
	Label string
	// Annotation filters documents by annotation string
	Annotation string
	// APIVersion filters documents by API group and version
	APIVersion string
	// Kind filters documents by document kind
	Kind string
}

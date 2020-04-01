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

import (
	"opendev.org/airship/airshipctl/pkg/environment"
)

// Settings for document rendering
type Settings struct {
	*environment.AirshipCTLSettings
	// Label filter documents by label string
	Label []string
	// Annotation filter documents by annotation string
	Annotation []string
	// GroupVersion filter documents by API version
	GroupVersion []string
	// Kind filter documents by document kind
	Kind []string
	// RawFilter contains logical expression to filter documents
	RawFilter string
}

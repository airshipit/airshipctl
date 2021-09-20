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

package version

var (
	gitVersion = "devel"
	gitCommit  = ""
	buildDate  = ""
)

// Info structure provides version data for airshipctl
// For now it has GitVersion with format 'v0.1.0'
// provided from Makefile during building airshipctl
// or defined in a local var for development purposes
type Info struct {
	GitVersion string `json:"gitVersion,omitempty"`
	GitCommit  string `json:"gitCommit,omitempty"`
	BuildDate  string `json:"buildDate,omitempty"`
}

// Get function shows airshipctl version
// returns filled Info structure
func Get() Info {
	return Info{
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
	}
}

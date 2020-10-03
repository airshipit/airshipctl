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

package isogen

// ErrIsoGenNilBundle is returned when isogen executor is not provided with bundle
type ErrIsoGenNilBundle struct {
}

func (e ErrIsoGenNilBundle) Error() string {
	return "Cannot build iso with empty bundle, no data source is available"
}

// ErrNoParsedNumPkgs is returned when it's unable to find number of packages to install
type ErrNoParsedNumPkgs struct {
}

func (e ErrNoParsedNumPkgs) Error() string {
	return "No number of packages to install found in parsed container logs"
}

// ErrUnexpectedPb is returned when progress bar was not finished for some reason
type ErrUnexpectedPb struct {
}

func (e ErrUnexpectedPb) Error() string {
	return "An unexpected error occurred while parsing container logs"
}

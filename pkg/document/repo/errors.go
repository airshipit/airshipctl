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

package repo

import (
	"fmt"
)

// ErrNoOpenRepo returned if repository is not opened by git driver
type ErrNoOpenRepo struct {
}

// ErrParseURL returned if there is error when parsing URL
type ErrParseURL struct {
	URL string
}

func (e ErrNoOpenRepo) Error() string {
	return fmt.Sprintf("no open repository is stored")
}

func (e ErrParseURL) Error() string {
	return fmt.Sprintf("could not get target directory from URL: %s", e.URL)
}

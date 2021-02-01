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

package extlib

import (
	"github.com/lucasjones/reggen"
)

// Generate Regex
func regexGen(regex string, limit int) string {
	if limit <= 0 {
		panic("Limit cannot be less than or equal to 0")
	}
	str, err := reggen.Generate(regex, limit)
	if err != nil {
		panic(err)
	}
	return str
}

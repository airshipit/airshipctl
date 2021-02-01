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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToYaml(t *testing.T) {
	data := []struct {
		A string
		B int
	}{
		{
			A: "test1",
			B: 1,
		},
		{
			A: "test2",
			B: 2,
		},
	}

	assert.Equal(t, `
- a: test1
  b: 1
- a: test2
  b: 2
`[1:], toYaml(&data))
}

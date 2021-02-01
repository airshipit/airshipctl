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

func TestRegexGen(t *testing.T) {
	tpl := `{{- $regex := "^[a-z]{5,10}$" }}
{{- $nomatchregex := "^[a-z]{0,4}$" }}
true={{- regexMatch $regex (regexGen $regex 10) }},
false={{- regexMatch $nomatchregex (regexGen $regex 10) }}
`
	out, err := runRaw(tpl, nil)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, `
true=true,
false=false
`, out)
}

func TestRegexPanicOnPattern(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	regexGen("[a-z", 1)
}

func TestRegexPanicOnLimit(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	regexGen("[a-z]{0,4}", 0)
}

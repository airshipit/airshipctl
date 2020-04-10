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

package util_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/util"
)

func TestNewTabWriter(t *testing.T) {
	var tests = []struct {
		testname, src, expected string
	}{
		{
			"empty-string",
			"",
			"\n",
		},
		{
			"newline-test",
			"\n",
			"\n\n",
		},
		{
			"format-string",
			"airshipctl\tconfig\tinit",
			"airshipctl   config   init\n",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.testname, func(t *testing.T) {
			var buf bytes.Buffer
			out := util.NewTabWriter(&buf)
			fmt.Fprintln(out, tt.src)
			err := out.Flush()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

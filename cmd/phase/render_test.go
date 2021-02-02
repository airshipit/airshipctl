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

package phase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/cmd/phase"
	pkgphase "opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRender(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "render-with-help",
			CmdLine: "-h",
			Cmd:     phase.NewRenderCommand(nil),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

func TestRenderArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedErr error
	}{
		{
			name:        "error 2 args",
			args:        []string{"my-phase", "accidental"},
			expectedErr: &phase.ErrRenderTooManyArgs{Count: 2},
		},
		{
			name: "success",
			args: []string{"my-phase"},
		},
		{
			name: "success no args",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := phase.RenderArgs(&pkgphase.RenderCommand{})(phase.NewRenderCommand(nil), tt.args)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

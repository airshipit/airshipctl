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

	"opendev.org/airship/airshipctl/cmd/phase"
	"opendev.org/airship/airshipctl/testutil"
)

func TestStatus(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "run-with-help",
			CmdLine: "-h",
			Cmd:     phase.NewStatusCommand(nil),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

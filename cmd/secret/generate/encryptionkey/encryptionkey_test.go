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

package encryptionkey_test

import (
	"fmt"
	"testing"

	"opendev.org/airship/airshipctl/cmd/secret/generate/encryptionkey"
	"opendev.org/airship/airshipctl/testutil"
)

func TestGenerateEncryptionKey(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "generate-encryptionkey-cmd-with-help",
			CmdLine: "--help",
			Cmd:     encryptionkey.NewGenerateEncryptionKeyCommand(),
		},
		{
			Name:    "generate-encryptionkey-cmd-error",
			CmdLine: "--limit 10",
			Error:   fmt.Errorf("required Regex flag with limit option"),
			Cmd:     encryptionkey.NewGenerateEncryptionKeyCommand(),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

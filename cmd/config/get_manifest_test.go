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

package config_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestNewGetManifestCommand(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-get-manifest-help",
			CmdLine: "-h",
			Cmd:     config.NewGetManifestCommand(nil),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestGetManifestNArgs(t *testing.T) {
	tests := []struct {
		name         string
		use          string
		args         []string
		manifestName string
		err          error
	}{
		{
			name: "get-manifests no args",
			use:  "get-manifests",
			args: []string{},
			err:  nil,
		},
		{
			name: "get-manifests 1 arg",
			use:  "get-manifests",
			args: []string{"arg1"},
			err:  fmt.Errorf("accepts 0 arg(s), received 1"),
		},
		{
			name: "get-manifest no args",
			use:  "get-manifest",
			args: []string{},
			err:  fmt.Errorf("accepts 1 arg(s), received 0"),
		},
		{
			name: "get-manifest 1 arg",
			use:  "get-manifest",
			args: []string{"arg1"},
			err:  nil,
		},
	}

	out := &bytes.Buffer{}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:  tt.use,
				Args: config.GetManifestNArgs(&tt.manifestName),
				Run:  func(cmd *cobra.Command, args []string) {},
			}
			cmd.SetArgs(tt.args)
			cmd.SetOut(out)
			err := cmd.Execute()
			if tt.err != nil {
				require.Equal(t, tt.err, err)
			} else if len(tt.args) == 1 {
				require.Equal(t, tt.args[0], tt.manifestName)
			}
			out.Reset()
		})
	}
}

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

package cmd_test

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/cmd/bootstrap"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRoot(t *testing.T) {
	tests := []*testutil.CmdTest{
		{
			Name:    "rootCmd-with-no-defaults",
			CmdLine: "-h",
			Cmd:     getVanillaRootCmd(t),
		},
		{
			Name:    "rootCmd-with-defaults",
			CmdLine: "-h",
			Cmd:     getDefaultRootCmd(t),
		},
		{
			Name:    "specialized-rootCmd-with-bootstrap",
			CmdLine: "-h",
			Cmd:     getSpecializedRootCmd(t),
		},
	}

	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

func TestFlagLoading(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		Error    error
	}{
		{
			name:     "default, no flags",
			args:     []string{},
			expected: "",
			Error:    cobra.ErrSubCommandRequired,
		},
		{
			name:     "alternate airshipconfig",
			args:     []string{"--airshipconf", "/custom/path/to/airshipconfig"},
			expected: "/custom/path/to/airshipconfig",
			Error:    cobra.ErrSubCommandRequired,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(subTest *testing.T) {
			// We don't care about the output of this test, so toss
			// it into a throwaway &bytes.buffer{}
			rootCmd, settings, err := cmd.NewRootCmd(&bytes.Buffer{})
			require.NoError(t, err)
			rootCmd.SetArgs(tt.args)

			err = rootCmd.Execute()
			assert.Equal(t, tt.Error, err)

			assert.Equal(t, settings.AirshipConfigPath, tt.expected)
		})
	}
}

func getVanillaRootCmd(t *testing.T) *cobra.Command {
	t.Helper()
	rootCmd, _, err := cmd.NewRootCmd(nil)
	require.NoError(t, err, "Could not create root commands")
	return rootCmd
}

func getDefaultRootCmd(t *testing.T) *cobra.Command {
	t.Helper()
	rootCmd, _, err := cmd.NewAirshipCTLCommand(nil)
	require.NoError(t, err, "Could not create root commands")
	return rootCmd
}

func getSpecializedRootCmd(t *testing.T) *cobra.Command {
	rootCmd := getVanillaRootCmd(t)
	rootCmd.AddCommand(bootstrap.NewBootstrapCommand(&environment.AirshipCTLSettings{}))
	return rootCmd
}

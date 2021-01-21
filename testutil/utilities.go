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

package testutil

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// UpdateGolden writes out the golden files with the latest values, rather than failing the test.
var shouldUpdateGolden = flag.Bool("update", false, "update golden files")

const (
	testdataDir      = "testdata"
	goldenDirSuffix  = "GoldenOutput"
	goldenFileSuffix = ".golden"
)

// CmdTest is a command to be run on the command line as a test
type CmdTest struct {
	// The name of the test. This will be used when generating golden
	// files
	Name string

	// The values that would be inputted to airshipctl as commands, flags,
	// and arguments. The initial "airshipctl" is implied
	CmdLine string

	// The instantiated version of the root airshipctl command to test
	Cmd *cobra.Command

	// The expected error
	Error error
}

// RunTest either asserts that a specific command's output matches the expected
// output from its golden file, or generates golden files if the -update flag
// is passed
func RunTest(t *testing.T, test *CmdTest) {
	t.Helper()
	cmd := test.Cmd

	actual := &bytes.Buffer{}
	cmd.SetOut(actual)

	args := strings.Fields(test.CmdLine)
	cmd.SetArgs(args)

	err := cmd.Execute()
	checkError(t, err, test.Error)

	if *shouldUpdateGolden {
		updateGolden(t, test, actual.Bytes())
	} else {
		assertEqualGolden(t, test, actual.Bytes())
	}
}

func updateGolden(t *testing.T, test *CmdTest, actual []byte) {
	t.Helper()
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	err := os.MkdirAll(goldenDir, 0775)
	require.NoErrorf(t, err, "Failed to create golden directory %s", goldenDir)
	t.Logf("Created %s", goldenDir)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	t.Logf("Updating golden file: %s", goldenFilePath)
	err = ioutil.WriteFile(goldenFilePath, normalize(actual), 0600)
	require.NoErrorf(t, err, "Failed to update golden file at %s", goldenFilePath)
}

func assertEqualGolden(t *testing.T, test *CmdTest, actual []byte) {
	t.Helper()
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	golden, err := ioutil.ReadFile(goldenFilePath)
	require.NoErrorf(t, err, "Failed while reading golden file at %s", goldenFilePath)
	assert.Equal(t, string(golden), string(actual))
}

func checkError(t *testing.T, actual, expected error) {
	t.Helper()
	if expected == nil {
		require.NoError(t, actual)
	} else {
		require.Error(t, actual)
		require.Equal(t, expected.Error(), actual.Error())
	}
}

func normalize(in []byte) []byte {
	return bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
}

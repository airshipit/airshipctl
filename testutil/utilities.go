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

	// The instatiated version of the root airshipctl command to test
	Cmd *cobra.Command

	// The expected error
	Error error
}

// RunTest either asserts that a specific command's output matches the expected
// output from its golden file, or generates golden files if the -update flag
// is passed
func RunTest(t *testing.T, test *CmdTest) {
	cmd := test.Cmd

	actual := &bytes.Buffer{}
	cmd.SetOutput(actual)

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

// ReadFixtureBytes is a convenience function for opening a test fixture
func ReadFixtureBytes(t *testing.T, filename string) []byte {
	fixtureData, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Unexpected error while reading fixture at %s: %s", filename, err.Error())
	}
	return fixtureData
}

// ReadFixtureString is a convenience function for opening a test fixture
func ReadFixtureString(t *testing.T, filename string) string {
	return string(ReadFixtureBytes(t, filename))
}

func updateGolden(t *testing.T, test *CmdTest, actual []byte) {
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	if err := os.MkdirAll(goldenDir, 0775); err != nil {
		t.Fatalf("Failed to create golden directory %s: %s", goldenDir, err)
	}
	t.Logf("Created %s", goldenDir)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	t.Logf("Updating golden file: %s", goldenFilePath)
	if err := ioutil.WriteFile(goldenFilePath, normalize(actual), 0666); err != nil {
		t.Fatalf("Failed to update golden file: %s", err)
	}
}

func assertEqualGolden(t *testing.T, test *CmdTest, actual []byte) {
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	golden, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("Failed while reading golden file: %s", err)
	}
	if !bytes.Equal(actual, golden) {
		errFmt := "Output does not match golden file: %s\nEXPECTED:\n%s\nGOT:\n%s"
		t.Errorf(errFmt, goldenFilePath, string(golden), string(actual))
	}
}

func checkError(t *testing.T, actual, expected error) {
	if expected == nil {
		if actual == nil {
			return
		}
		t.Fatalf("Unexpected error: %q", actual.Error())
	}

	if actual == nil {
		t.Fatalf("Expected error %q, but got nil", expected.Error())
	}

	if actual.Error() != expected.Error() {
		t.Fatalf("Expected error %q, but got %q", expected.Error(), actual.Error())
	}
}

func normalize(in []byte) []byte {
	return bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
}

package test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ian-howell/airshipctl/pkg/util"
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
	Name    string
	CmdLine string
	Objs    []runtime.Object
}

// RunTest either asserts that a specific command's output matches the expected
// output from its golden file, or generates golden files if the -update flag
// is passed
func RunTest(t *testing.T, test *CmdTest, cmd *cobra.Command) {
	util.InitClock()
	actual := &bytes.Buffer{}
	cmd.SetOutput(actual)
	args := strings.Fields(test.CmdLine)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}
	if *shouldUpdateGolden {
		updateGolden(t, test, actual.Bytes())
	} else {
		assertEqualGolden(t, test, actual.Bytes())
	}
}

func updateGolden(t *testing.T, test *CmdTest, actual []byte) {
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	if err := os.MkdirAll(goldenDir, 0775); err != nil {
		t.Fatalf("failed to create golden directory %s: %s", goldenDir, err)
	}
	t.Logf("Created %s", goldenDir)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	t.Logf("updating golden file: %s", goldenFilePath)
	if err := ioutil.WriteFile(goldenFilePath, normalize(actual), 0666); err != nil {
		t.Fatalf("failed to update golden file: %s", err)
	}
}

func assertEqualGolden(t *testing.T, test *CmdTest, actual []byte) {
	goldenDir := filepath.Join(testdataDir, t.Name()+goldenDirSuffix)
	goldenFilePath := filepath.Join(goldenDir, test.Name+goldenFileSuffix)
	golden, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("failed while reading golden file: %s", err)
	}
	if !bytes.Equal(actual, golden) {
		errFmt := "output does not match golden file: %s\nEXPECTED:\n%s\nGOT:\n%s"
		t.Errorf(errFmt, goldenFilePath, string(golden), string(actual))
	}
}

func normalize(in []byte) []byte {
	return bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
}

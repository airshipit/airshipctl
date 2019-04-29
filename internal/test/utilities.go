package test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ian-howell/airshipadm/cmd"
)

// UpdateGolden writes out the golden files with the latest values, rather than failing the test.
var updateGolden = flag.Bool("update", false, "update golden files")

const goldenFileDir = "testdata/golden"

type CmdTest struct {
	Name    string
	Command string
}

func RunCmdTests(t *testing.T, tests []CmdTest) {
	t.Helper()

	for _, test := range tests {
		executeCmd(t, test.Command)
	}
}

func executeCmd(t *testing.T, command string) {
	var actual bytes.Buffer
	rootCmd := cmd.NewRootCmd(&actual)

	// TODO(howell): switch to shellwords (or similar)
	args := strings.Fields(command)
	rootCmd.SetArgs(args)

	rootCmd.Execute()

	goldenFilePath := filepath.Join(goldenFileDir, filepath.FromSlash(t.Name())+".golden")
	if *updateGolden {
		t.Logf("updating golden file: %s", goldenFilePath)
		if err := ioutil.WriteFile(goldenFilePath, actual.Bytes(), 0644); err != nil {
			t.Fatalf("failed to update golden file: %s", err)
		}
		return
	}
	golden, err := ioutil.ReadFile(goldenFilePath)
	if err != nil {
		t.Fatalf("failed while reading golden file: %s", err)
	}
	if !bytes.Equal(actual.Bytes(), golden) {
		errFmt := "output does not match golden file: %s\nEXPECTED:\n%s\nGOT:\n%s"
		t.Errorf(errFmt, goldenFilePath, string(golden), actual.String)
	}
}

func normalize(in []byte) []byte {
	return bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
}

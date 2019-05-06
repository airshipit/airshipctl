package test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ian-howell/airshipadm/cmd"
	"github.com/ian-howell/airshipadm/pkg/kube"
	"k8s.io/client-go/kubernetes/fake"
)

// UpdateGolden writes out the golden files with the latest values, rather than failing the test.
var shouldUpdateGolden = flag.Bool("update", false, "update golden files")

const (
	testdataDir      = "testdata"
	goldenDirSuffix  = "GoldenOutput"
	goldenFileSuffix = ".golden"
)

type CmdTest struct {
	Name    string
	Command string
}

func RunCmdTests(t *testing.T, tests []CmdTest) {
	t.Helper()
	for _, test := range tests {
		cmdOutput := executeCmd(t, test.Command)
		if *shouldUpdateGolden {
			updateGolden(t, test, cmdOutput)
		} else {
			assertEqualGolden(t, test, cmdOutput)
		}
	}
}

func executeCmd(t *testing.T, command string) []byte {
	var actual bytes.Buffer
	client := &kube.Client{Interface: fake.NewSimpleClientset()}
	// TODO(howell): switch to shellwords (or similar)
	args := strings.Fields(command)
	rootCmd, err := cmd.NewRootCmd(&actual, client, args)
	if err != nil {
		t.Fatalf(err.Error())
	}

	rootCmd.SetArgs(args)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf(err.Error())
	}

	return actual.Bytes()
}

func updateGolden(t *testing.T, test CmdTest, actual []byte) {
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

func assertEqualGolden(t *testing.T, test CmdTest, actual []byte) {
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

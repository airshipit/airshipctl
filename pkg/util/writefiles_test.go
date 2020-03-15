package util_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/testutil"
)

func TestWriteFiles(t *testing.T) {
	testDir, cleanup := testutil.TempDir(t, "test-dir")
	defer cleanup(t)

	fls := make(map[string][]byte)
	dummyData := []byte("")
	testFile1 := filepath.Join(testDir, "testFile1")
	testFile2 := filepath.Join(testDir, "testFile2")
	fls[testFile1] = dummyData
	fls[testFile2] = dummyData
	err := util.WriteFiles(fls, 0600)

	assert.NoError(t, err)

	// check if all files are created
	assert.FileExists(t, testFile1)
	assert.FileExists(t, testFile2)

	// check if files are readable
	_, err = ioutil.ReadFile(testFile1)
	assert.NoError(t, err)
}

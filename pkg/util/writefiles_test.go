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

	// test to fail WriteFiles with NonExistent Dir Path
	testFile3 := filepath.Join("NonExistentDir", "testFile3")
	fls[testFile3] = dummyData
	err = util.WriteFiles(fls, 0600)
	assert.Error(t, err)
}

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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	mName       = "dummymanifest"
	mRepoName   = "treasuremap"
	mURL        = "https://github.com/airshipit/treasuremap"
	mBranch     = "master"
	mSubPath    = "manifests/test-site"
	mTargetPath = "/tmp/airship"

	testTargetPath = "/tmp/e2e"
)

type setManifestTest struct {
	inputConfig        *config.Config
	cmdTest            *testutil.CmdTest
	manifestName       string
	manifestTargetPath string
}

func TestConfigSetManifest(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-manifest-with-help",
			CmdLine: "--help",
			Cmd:     cmd.NewSetManifestCommand(nil),
		},
		{
			Name:    "config-cmd-set-manifest-too-many-args",
			CmdLine: "arg1 arg2",
			Cmd:     cmd.NewSetManifestCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 2),
		},
		{
			Name:    "config-cmd-set-manifest-too-few-args",
			CmdLine: "",
			Cmd:     cmd.NewSetManifestCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 0),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestSetManifest(t *testing.T) {
	given, cleanupGiven := testutil.InitConfig(t)
	defer cleanupGiven(t)

	tests := []struct {
		testName     string
		manifestName string
		flags        []string
		givenConfig  *config.Config
		targetPath   string
	}{
		{
			testName:     "set-manifest",
			manifestName: mName,
			flags: []string{
				"--repo " + mRepoName,
				"--url " + mURL,
				"--branch " + mBranch,
				"--primary",
				"--sub-path " + mSubPath,
				"--target-path " + mTargetPath,
			},
			givenConfig: given,
			targetPath:  mTargetPath,
		},
		{
			testName:     "modify-manifest",
			manifestName: mName,
			flags: []string{
				"--target-path " + testTargetPath,
			},
			givenConfig: given,
			targetPath:  testTargetPath,
		},
	}

	for _, tt := range tests {
		tt := tt
		cmd := &testutil.CmdTest{
			Name:    tt.testName,
			CmdLine: fmt.Sprintf("%s %s", tt.manifestName, strings.Join(tt.flags, " ")),
		}
		test := setManifestTest{
			inputConfig:        tt.givenConfig,
			cmdTest:            cmd,
			manifestName:       tt.manifestName,
			manifestTargetPath: tt.targetPath,
		}
		test.run(t)
	}
}

func (test setManifestTest) run(t *testing.T) {
	settings := &environment.AirshipCTLSettings{Config: test.inputConfig}
	test.cmdTest.Cmd = cmd.NewSetManifestCommand(settings)
	testutil.RunTest(t, test.cmdTest)

	afterRunConf := settings.Config
	// Find the Manifest Created or Modified
	afterRunManifest, _ := afterRunConf.Manifests[test.manifestName]
	require.NotNil(t, afterRunManifest)
	assert.EqualValues(t, afterRunManifest.TargetPath, test.manifestTargetPath)
}

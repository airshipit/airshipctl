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

package pull

import (
	"io/ioutil"
	"path"
	"strings"
	"testing"

	fixtures "gopkg.in/src-d/go-git-fixtures.v3"

	repo2 "opendev.org/airship/airshipctl/pkg/document/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/testutil"
)

func getDummyPullSettings() *Settings {
	mockPullSettings := &Settings{
		AirshipCTLSettings: new(environment.AirshipCTLSettings),
	}
	mockConf := testutil.DummyConfig()
	mockPullSettings.AirshipCTLSettings.SetConfig(mockConf)
	return mockPullSettings
}

func TestPull(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	t.Run("cloneRepositories", func(t *testing.T) {
		dummyPullSettings := getDummyPullSettings()
		currentManifest, err := dummyPullSettings.Config().CurrentContextManifest()
		require.NoError(err)

		err = fixtures.Init()
		require.NoError(err)
		fx := fixtures.Basic().One()

		dummyGitDir := fx.DotGit().Root()
		currentManifest.Repositories = map[string]*config.Repository{currentManifest.PrimaryRepositoryName: {
			URLString: dummyGitDir,
			CheckoutOptions: &config.RepoCheckout{
				Branch:        "master",
				ForceCheckout: false,
			},
			Auth: &config.RepoAuth{
				Type: "http-basic",
			},
		},
		}

		tmpDir, cleanup := testutil.TempDir(t, "airshipctlPullTest-")
		defer cleanup(t)

		currentManifest.TargetPath = tmpDir

		_, err = repo2.NewRepository(".", currentManifest.Repositories[currentManifest.PrimaryRepositoryName])
		require.NoError(err)

		err = dummyPullSettings.cloneRepositories()

		require.NoError(err)
		dummyRepoDirName := util.GitDirNameFromURL(dummyGitDir)
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, "go/example.go"))
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		contents, err := ioutil.ReadFile(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		require.NoError(err)
		assert.Equal("ref: refs/heads/master", strings.TrimRight(string(contents), "\t \n"))
	})

	t.Run("Pull", func(t *testing.T) {
		dummyPullSettings := getDummyPullSettings()
		conf := dummyPullSettings.AirshipCTLSettings.Config()

		err := fixtures.Init()
		require.NoError(err)
		fx := fixtures.Basic().One()

		mfst := conf.Manifests["dummy_manifest"]
		dummyGitDir := fx.DotGit().Root()
		mfst.Repositories = map[string]*config.Repository{
			mfst.PrimaryRepositoryName: {
				URLString: dummyGitDir,
				CheckoutOptions: &config.RepoCheckout{
					Branch:        "master",
					ForceCheckout: false,
				},
				Auth: &config.RepoAuth{
					Type: "http-basic",
				},
			},
		}
		dummyPullSettings.SetConfig(conf)

		tmpDir, cleanup := testutil.TempDir(t, "airshipctlPullTest-")
		defer cleanup(t)

		mfst.TargetPath = tmpDir
		require.NoError(err)

		err = dummyPullSettings.Pull()
		require.NoError(err)

		dummyRepoDirName := util.GitDirNameFromURL(dummyGitDir)
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, "go/example.go"))
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		contents, err := ioutil.ReadFile(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		require.NoError(err)
		assert.Equal("ref: refs/heads/master", strings.TrimRight(string(contents), "\t \n"))
	})

	testutil.CleanUpGitFixtures(t)
}

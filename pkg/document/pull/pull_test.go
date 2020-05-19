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

	fixtures "github.com/go-git/go-git-fixtures/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document/repo"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/util"
	"opendev.org/airship/airshipctl/testutil"
)

func getDummyPullSettings() *Settings {
	return &Settings{
		AirshipCTLSettings: &environment.AirshipCTLSettings{
			Config: testutil.DummyConfig(),
		},
	}
}

func TestPull(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	tests := []struct {
		name         string
		url          string
		checkoutOpts *config.RepoCheckout
		error        error
	}{
		{
			name: "TestCloneRepositoriesValidOpts",
			checkoutOpts: &config.RepoCheckout{
				Branch:        "master",
				ForceCheckout: false,
			},
			error: nil,
		},
		{
			name:  "TestCloneRepositoriesMissingCheckoutOptions",
			error: nil,
		},
		{
			name: "TestCloneRepositoriesNonMasterBranch",
			checkoutOpts: &config.RepoCheckout{
				Branch:        "branch",
				ForceCheckout: false,
			},
			error: nil,
		},
		{
			name: "TestCloneRepositoriesInvalidOpts",
			checkoutOpts: &config.RepoCheckout{
				Branch:        "master",
				Tag:           "someTag",
				ForceCheckout: false,
			},
			error: config.ErrMutuallyExclusiveCheckout{},
		},
	}
	dummyPullSettings := getDummyPullSettings()
	currentManifest, err := dummyPullSettings.Config.CurrentContextManifest()
	require.NoError(err)

	testGitDir := fixtures.Basic().One().DotGit().Root()
	dirNameFromURL := util.GitDirNameFromURL(testGitDir)
	globalTmpDir, cleanup := testutil.TempDir(t, "airshipctlCloneTest-")
	defer cleanup(t)

	for _, tt := range tests {
		tmpDir := path.Join(globalTmpDir, tt.name)
		expectedErr := tt.error
		chkOutOpts := tt.checkoutOpts
		t.Run(tt.name, func(t *testing.T) {
			currentManifest.Repositories = map[string]*config.Repository{
				currentManifest.PrimaryRepositoryName: {
					URLString:       testGitDir,
					CheckoutOptions: chkOutOpts,
					Auth: &config.RepoAuth{
						Type: "http-basic",
					},
				},
			}

			currentManifest.TargetPath = tmpDir

			_, err = repo.NewRepository(
				".",
				currentManifest.Repositories[currentManifest.PrimaryRepositoryName],
			)
			require.NoError(err)

			err = dummyPullSettings.Pull()
			if expectedErr != nil {
				assert.NotNil(err)
				assert.Equal(expectedErr, err)
			} else {
				require.NoError(err)
				assert.FileExists(path.Join(currentManifest.TargetPath, dirNameFromURL, "go/example.go"))
				assert.FileExists(path.Join(currentManifest.TargetPath, dirNameFromURL, ".git/HEAD"))
				contents, err := ioutil.ReadFile(path.Join(currentManifest.TargetPath, dirNameFromURL, ".git/HEAD"))
				require.NoError(err)
				if chkOutOpts == nil {
					assert.Equal(
						"ref: refs/heads/master",
						strings.TrimRight(string(contents), "\t \n"),
					)
				} else {
					assert.Equal(
						"ref: refs/heads/"+chkOutOpts.Branch,
						strings.TrimRight(string(contents), "\t \n"),
					)
				}
			}
		})
	}
	testutil.CleanUpGitFixtures(t)
}

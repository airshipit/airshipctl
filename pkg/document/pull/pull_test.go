package pull

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"testing"

	fixtures "gopkg.in/src-d/go-git-fixtures.v3"

	repo2 "opendev.org/airship/airshipctl/pkg/document/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

func getDummyPullSettings() *Settings {
	mockPullSettings := &Settings{
		AirshipCTLSettings: new(environment.AirshipCTLSettings),
	}
	mockConf := config.DummyConfig()
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
		currentManifest.Repository = &config.Repository{
			URLString: dummyGitDir,
			CheckoutOptions: &config.RepoCheckout{
				Branch:        "master",
				ForceCheckout: false,
			},
			Auth: &config.RepoAuth{
				Type: "http-basic",
			},
		}

		tmpDir, err := ioutil.TempDir("", "airshipctlPullTest-")
		require.NoError(err)
		currentManifest.TargetPath = tmpDir

		_, err = repo2.NewRepository(".", currentManifest.Repository)
		require.NoError(err)

		err = dummyPullSettings.cloneRepositories()

		require.NoError(err)
		dummyRepoDirName := filepath.Base(dummyGitDir)
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
		mfst.Repository = &config.Repository{
			URLString: dummyGitDir,
			CheckoutOptions: &config.RepoCheckout{
				Branch:        "master",
				ForceCheckout: false,
			},
			Auth: &config.RepoAuth{
				Type: "http-basic",
			},
		}
		dummyPullSettings.SetConfig(conf)

		tmpDir, err := ioutil.TempDir("", "airshipctlPullTest-")
		require.NoError(err)
		mfst.TargetPath = tmpDir
		require.NoError(err)

		err = dummyPullSettings.Pull()
		require.NoError(err)

		dummyRepoDirName := filepath.Base(dummyGitDir)
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, "go/example.go"))
		assert.FileExists(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		contents, err := ioutil.ReadFile(path.Join(tmpDir, dummyRepoDirName, ".git/HEAD"))
		require.NoError(err)
		assert.Equal("ref: refs/heads/master", strings.TrimRight(string(contents), "\t \n"))
	})
}

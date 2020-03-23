package document

import (
	"testing"

	"github.com/stretchr/testify/require"
	fixtures "gopkg.in/src-d/go-git-fixtures.v3"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"

	"opendev.org/airship/airshipctl/testutil"
)

func getDummyAirshipSettings(t *testing.T) *environment.AirshipCTLSettings {
	settings := &environment.AirshipCTLSettings{Config: testutil.DummyConfig()}

	err := fixtures.Init()
	require.NoError(t, err)

	fx := fixtures.Basic().One()

	mfst := settings.Config.Manifests["dummy_manifest"]
	mfst.Repositories = map[string]*config.Repository{
		"primary": {
			URLString: fx.DotGit().Root(),
			CheckoutOptions: &config.RepoCheckout{
				Branch:        "master",
				ForceCheckout: false,
			},
			Auth: &config.RepoAuth{
				Type: "http-basic",
			},
		},
	}
	return settings
}

func TestPull(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "document-pull-cmd-with-defaults",
			CmdLine: "",
			Cmd:     NewDocumentPullCommand(getDummyAirshipSettings(t)),
		},
		{
			Name:    "document-pull-cmd-with-help",
			CmdLine: "--help",
			Cmd:     NewDocumentPullCommand(nil),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}

	testutil.CleanUpGitFixtures(t)
}

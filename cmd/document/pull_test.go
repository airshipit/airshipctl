package document

import (
	"testing"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"

	"opendev.org/airship/airshipctl/testutil"
)

func getDummyAirshipSettings() *environment.AirshipCTLSettings {
	settings := new(environment.AirshipCTLSettings)
	conf := config.DummyConfig()
	mfst := conf.Manifests["dummy_manifest"]
	mfst.Repository = &config.Repository{
		URLString: "https://opendev.org/airship/treasuremap.git",
		CheckoutOptions: &config.RepoCheckout{
			Branch:        "master",
			ForceCheckout: false,
		},
		Auth: &config.RepoAuth{
			Type: "http-basic",
		},
	}
	settings.SetConfig(conf)
	return settings
}

func TestPull(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "document-pull-cmd-with-defaults",
			CmdLine: "",
			Cmd:     NewDocumentPullCommand(getDummyAirshipSettings()),
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
}

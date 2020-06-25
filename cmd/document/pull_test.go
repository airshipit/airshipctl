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

package document

import (
	"testing"

	fixtures "github.com/go-git/go-git-fixtures/v4"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func getDummyAirshipSettings() *environment.AirshipCTLSettings {
	settings := &environment.AirshipCTLSettings{Config: testutil.DummyConfig()}

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
			Name:    "document-pull-cmd",
			CmdLine: "",
			Cmd:     NewPullCommand(getDummyAirshipSettings(), false),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}

	testutil.CleanUpGitFixtures(t)
}

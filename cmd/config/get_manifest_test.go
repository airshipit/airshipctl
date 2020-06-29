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
	"testing"

	cmd "opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestGetManifestConfigCmd(t *testing.T) {
	settings := &environment.AirshipCTLSettings{
		Config: &config.Config{
			Manifests: map[string]*config.Manifest{
				"fooManifestConfig": getTestManifest("foo"),
				"barManifestConfig": getTestManifest("bar"),
				"bazManifestConfig": getTestManifest("baz"),
			},
		},
	}

	noConfigSettings := &environment.AirshipCTLSettings{Config: new(config.Config)}

	cmdTests := []*testutil.CmdTest{
		{
			Name:    "get-manifest",
			CmdLine: "fooManifestConfig",
			Cmd:     cmd.NewGetManifestCommand(settings),
		},
		{
			Name:    "get-all-manifests",
			CmdLine: "",
			Cmd:     cmd.NewGetManifestCommand(settings),
		},
		{
			Name:    "missing",
			CmdLine: "manifestMissing",
			Cmd:     cmd.NewGetManifestCommand(settings),
			Error:   fmt.Errorf("Missing configuration: Manifest with name 'manifestMissing'"),
		},
		{
			Name:    "no-manifests",
			CmdLine: "",
			Cmd:     cmd.NewGetManifestCommand(noConfigSettings),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func getTestManifest(name string) *config.Manifest {
	manifests := &config.Manifest{
		PrimaryRepositoryName: name + "_primary_repo",
		TargetPath:            name + "_target_path",
	}
	return manifests
}

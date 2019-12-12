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

package phase_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/cmd/phase"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRender(t *testing.T) {
	cfg, cleanupCfg := testutil.InitConfig(t)
	defer cleanupCfg(t)
	cfg.CurrentContext = "def_ephemeral"
	cfg.Manifests["test"] = &config.Manifest{
		TargetPath:            "testdata",
		PrimaryRepositoryName: "testRepo",
		Repositories: map[string]*config.Repository{
			"testRepo": {
				URLString: "http://localhost",
			},
		},
	}
	ctx, err := cfg.GetContext("def_ephemeral")
	require.NoError(t, err)
	ctx.Manifest = "test"
	settings := &environment.AirshipCTLSettings{Config: cfg}

	tests := []*testutil.CmdTest{
		{
			Name:    "render-with-help",
			CmdLine: "-h",
			Cmd:     phase.NewRenderCommand(nil),
		},
		{
			Name:    "render-with-multiple-labels",
			CmdLine: "initinfra -l app=helm,name=tiller",
			Cmd:     phase.NewRenderCommand(settings),
		},
	}
	for _, tt := range tests {
		testutil.RunTest(t, tt)
	}
}

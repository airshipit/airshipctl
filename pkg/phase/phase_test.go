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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/phase"
)

func TestPhasePlan(t *testing.T) {
	testCases := []struct {
		name         string
		settings     func() *environment.AirshipCTLSettings
		expectedPlan map[string][]string
		expectedErr  error
	}{
		{
			name: "No context",
			settings: func() *environment.AirshipCTLSettings {
				s := makeDefaultSettings()
				s.Config.CurrentContext = "badCtx"
				return s
			},
			expectedErr: config.ErrMissingConfig{What: "Context with name 'badCtx'"},
		},
		{
			name:     "Valid Phase Plan",
			settings: makeDefaultSettings,
			expectedPlan: map[string][]string{
				"group1": {
					"isogen",
					"remotedirect",
					"initinfra",
				},
			},
		},
		{
			name: "No Phase Plan",
			settings: func() *environment.AirshipCTLSettings {
				s := makeDefaultSettings()
				m, err := s.Config.CurrentContextManifest()
				require.NoError(t, err)
				m.SubPath = "no_plan_site"
				return s
			},
			expectedErr: document.ErrDocNotFound{
				Selector: document.Selector{
					Selector: types.Selector{
						Gvk: resid.Gvk{
							Group:   "airshipit.org",
							Version: "v1alpha1",
							Kind:    "PhasePlan",
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.Cmd{AirshipCTLSettings: tt.settings()}
			actualPlan, actualErr := cmd.Plan()
			assert.Equal(t, tt.expectedErr, actualErr)
			assert.Equal(t, tt.expectedPlan, actualPlan)
		})
	}
}

func makeDefaultSettings() *environment.AirshipCTLSettings {
	testSettings := &environment.AirshipCTLSettings{
		AirshipConfigPath: "testdata/airshipconfig.yaml",
		KubeConfigPath:    "testdata/kubeconfig.yaml",
	}
	testSettings.InitConfig()
	return testSettings
}

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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"

	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
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
					"some_phase",
					"capi_init",
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

func TestGetPhase(t *testing.T) {
	testCases := []struct {
		name          string
		settings      func() *environment.AirshipCTLSettings
		phaseName     string
		expectedPhase *airshipv1.Phase
		expectedErr   error
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
			name:      "Get existing phase",
			settings:  makeDefaultSettings,
			phaseName: "capi_init",
			expectedPhase: &airshipv1.Phase{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "airshipit.org/v1alpha1",
					Kind:       "Phase",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "capi_init",
				},
				Config: airshipv1.PhaseConfig{
					ExecutorRef: &corev1.ObjectReference{
						Kind:       "Clusterctl",
						APIVersion: "airshipit.org/v1alpha1",
						Name:       "clusterctl-v1",
					},
					DocumentEntryPoint: "manifests/site/test-site/auth",
				},
			},
		},
		{
			name:      "Get non-existing phase",
			settings:  makeDefaultSettings,
			phaseName: "some_name",
			expectedErr: document.ErrDocNotFound{
				Selector: document.Selector{
					Selector: types.Selector{
						Gvk: resid.Gvk{
							Group:   "airshipit.org",
							Version: "v1alpha1",
							Kind:    "Phase",
						},
						Name: "some_name",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.Cmd{AirshipCTLSettings: tt.settings()}
			actualPhase, actualErr := cmd.GetPhase(tt.phaseName)
			assert.Equal(t, tt.expectedErr, actualErr)
			assert.Equal(t, tt.expectedPhase, actualPhase)
		})
	}
}

func TestGetExecutor(t *testing.T) {
	testCases := []struct {
		name        string
		settings    func() *environment.AirshipCTLSettings
		phase       *airshipv1.Phase
		expectedExc ifc.Executor
		expectedErr error
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
			name:     "Get non-existing executor",
			settings: makeDefaultSettings,
			phase: &airshipv1.Phase{
				Config: airshipv1.PhaseConfig{
					ExecutorRef: &corev1.ObjectReference{
						APIVersion: "example.com/v1",
						Kind:       "SomeKind",
					},
				},
			},
			expectedErr: document.ErrDocNotFound{
				Selector: document.Selector{
					Selector: types.Selector{
						Gvk: resid.Gvk{
							Group:   "example.com",
							Version: "v1",
							Kind:    "SomeKind",
						},
					},
				},
			},
		},
		{
			name:     "Get unregistered executor",
			settings: makeDefaultSettings,
			phase: &airshipv1.Phase{
				Config: airshipv1.PhaseConfig{
					ExecutorRef: &corev1.ObjectReference{
						APIVersion: "airshipit.org/v1alpha1",
						Kind:       "SomeExecutor",
						Name:       "executor-name",
					},
					DocumentEntryPoint: "valid_site/phases",
				},
			},
			expectedErr: phase.ErrExecutorNotFound{
				GVK: schema.GroupVersionKind{
					Group:   "airshipit.org",
					Version: "v1alpha1",
					Kind:    "SomeExecutor",
				},
			},
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.Cmd{AirshipCTLSettings: tt.settings()}
			actualExc, actualErr := cmd.GetExecutor(tt.phase)
			assert.Equal(t, tt.expectedErr, actualErr)
			assert.Equal(t, tt.expectedExc, actualExc)
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

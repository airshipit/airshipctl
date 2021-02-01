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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	airshipv1 "opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	brokenMetaPath = "broken_metadata.yaml"
	noPlanMetaPath = "no_plan_site/metadata.yaml"
)

func TestHelperPhase(t *testing.T) {
	testCases := []struct {
		name        string
		errContains string

		phaseID       ifc.ID
		config        func(t *testing.T) *config.Config
		expectedPhase *airshipv1.Phase
	}{
		{
			name:    "Success Get existing phase",
			config:  testConfig,
			phaseID: ifc.ID{Name: "capi_init"},
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
					DocumentEntryPoint: "valid_site/phases",
				},
			},
		},
		{
			name:        "Error get non-existing phase",
			config:      testConfig,
			phaseID:     ifc.ID{Name: "some_name"},
			errContains: "found no documents",
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualPhase, actualErr := helper.Phase(tt.phaseID)
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, tt.expectedPhase, actualPhase)
			}
		})
	}
}

func TestHelperPlan(t *testing.T) {
	testPlanName := "phasePlan"
	testCases := []struct {
		name         string
		errContains  string
		expectedPlan *v1alpha1.PhasePlan
		config       func(t *testing.T) *config.Config
	}{
		{
			name:   "Valid Phase Plan",
			config: testConfig,
			expectedPlan: &airshipv1.PhasePlan{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PhasePlan",
					APIVersion: "airshipit.org/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: testPlanName,
				},
				Phases: []airshipv1.PhaseStep{
					{
						Name: "isogen",
					},
					{
						Name: "remotedirect",
					},
					{
						Name: "initinfra",
					},
					{
						Name: "some_phase",
					},
					{
						Name: "capi_init",
					},
				},
			},
		},
		{
			name: "No Phase Plan",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = noPlanMetaPath
				return conf
			},
			errContains: "found no documents",
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualPlan, actualErr := helper.Plan(ifc.ID{Name: testPlanName})
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, tt.expectedPlan, actualPlan)
			}
		})
	}
}

func TestHelperListPhases(t *testing.T) {
	testCases := []struct {
		name        string
		errContains string
		phaseLen    int
		config      func(t *testing.T) *config.Config
	}{
		{
			name:     "Success phase list",
			phaseLen: 5,
			config:   testConfig,
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
		{
			name: "Success 0 phases",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = noPlanMetaPath
				return conf
			},
			phaseLen: 0,
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualList, actualErr := helper.ListPhases()
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Len(t, actualList, tt.phaseLen)
			}
		})
	}
}

func TestHelperListPlans(t *testing.T) {
	testCases := []struct {
		name        string
		errContains string
		expectedLen int
		config      func(t *testing.T) *config.Config
	}{
		{
			name:        "Success plan list",
			expectedLen: 5,
			config:      testConfig,
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
		{
			name: "Success 0 plans",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = noPlanMetaPath
				return conf
			},
			expectedLen: 0,
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualList, actualErr := helper.ListPlans()
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Len(t, actualList, tt.expectedLen)
			}
		})
	}
}

func TestHelperClusterMapAPI(t *testing.T) {
	testCases := []struct {
		name         string
		errContains  string
		expectedCMap *v1alpha1.ClusterMap
		config       func(t *testing.T) *config.Config
	}{
		{
			name: "Success cluster map",
			expectedCMap: &airshipv1.ClusterMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "airshipit.org/v1alpha1",
					Kind:       "ClusterMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterctl-v1",
				},
				Map: map[string]*airshipv1.Cluster{
					"target": {
						Parent:            "ephemeral",
						DynamicKubeConfig: false,
					},
					"ephemeral": {},
				},
			},
			config: testConfig,
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
		{
			name: "Error no cluster map",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = noPlanMetaPath
				return conf
			},
			errContains: "found no documents",
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualCMap, actualErr := helper.ClusterMapAPIobj()
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.Equal(t, tt.expectedCMap, actualCMap)
			}
		})
	}
}

func TestHelperClusterMap(t *testing.T) {
	testCases := []struct {
		name        string
		errContains string
		config      func(t *testing.T) *config.Config
	}{
		{
			name:   "Success phase list",
			config: testConfig,
		},
		{
			name: "Error no cluster map",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = noPlanMetaPath
				return conf
			},
			errContains: "found no documents",
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)
			require.NotNil(t, helper)

			actualCMap, actualErr := helper.ClusterMap()
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				assert.NotNil(t, actualCMap)
			}
		})
	}
}

func TestHelperExecutorDoc(t *testing.T) {
	testCases := []struct {
		name             string
		errContains      string
		expectedExecutor string

		phaseID ifc.ID
		config  func(t *testing.T) *config.Config
	}{
		{
			name:             "Success Get existing phase",
			config:           testConfig,
			phaseID:          ifc.ID{Name: "capi_init"},
			expectedExecutor: "clusterctl-v1",
		},
		{
			name:        "Error get non-existing phase",
			config:      testConfig,
			phaseID:     ifc.ID{Name: "some_name"},
			errContains: "found no documents",
		},
		{
			name: "Error bundle path doesn't exist",
			config: func(t *testing.T) *config.Config {
				conf := testConfig(t)
				conf.Manifests["dummy_manifest"].MetadataPath = brokenMetaPath
				return conf
			},
			errContains: "no such file or directory",
		},
		{
			name:    "Error get phase without executor",
			config:  testConfig,
			phaseID: ifc.ID{Name: "no_executor_phase"},
			errContains: phase.ErrExecutorRefNotDefined{
				PhaseName: "no_executor_phase",
			}.Error(),
		},
	}

	for _, test := range testCases {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			helper, err := phase.NewHelper(tt.config(t))
			require.NoError(t, err)

			actualDoc, actualErr := helper.ExecutorDoc(tt.phaseID)
			if tt.errContains != "" {
				require.Error(t, actualErr)
				assert.Contains(t, actualErr.Error(), tt.errContains)
			} else {
				require.NoError(t, actualErr)
				require.NotNil(t, actualDoc)
				assert.Equal(t, tt.expectedExecutor, actualDoc.GetName())
			}
		})
	}
}

func TestHelperTargetPath(t *testing.T) {
	helper, err := phase.NewHelper(testConfig(t))
	require.NoError(t, err)
	require.NotNil(t, helper)
	assert.Equal(t, "testdata", helper.TargetPath())
}

func TestHelperPhaseBundleRoot(t *testing.T) {
	helper, err := phase.NewHelper(testConfig(t))
	require.NoError(t, err)
	require.NotNil(t, helper)
	expectedPhaseBundleRoot := filepath.Join("testdata", "valid_site", "phases")
	assert.Equal(t, expectedPhaseBundleRoot, helper.PhaseBundleRoot())
}

func TestHelperPhaseEntryPointBasePath(t *testing.T) {
	cfg := testConfig(t)
	cfg.Manifests["dummy_manifest"].TargetPath = "testdata"
	cfg.Manifests["dummy_manifest"].Repositories["primary"].URLString = "http://dummy.org/reponame.git"
	cfg.Manifests["dummy_manifest"].MetadataPath = "../valid_site_with_doc_prefix/metadata.yaml"
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	expectedPhaseEntryPointBasePath := filepath.Join("testdata", "reponame", "valid_site_with_doc_prefix", "phases")
	assert.Equal(t, expectedPhaseEntryPointBasePath, helper.PhaseEntryPointBasePath())
}

func TestHelperPhaseRepoDir(t *testing.T) {
	cfg := testConfig(t)
	cfg.Manifests["dummy_manifest"].Repositories["primary"].URLString = "http://dummy.org/reponame.git"
	cfg.Manifests["dummy_manifest"].MetadataPath = "../valid_site/metadata.yaml"
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	assert.Equal(t, "reponame", helper.PhaseRepoDir())
}

func TestHelperDocEntryPointPrefix(t *testing.T) {
	cfg := testConfig(t)
	cfg.Manifests["dummy_manifest"].MetadataPath = "valid_site_with_doc_prefix/metadata.yaml"
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	assert.Equal(t, "valid_site_with_doc_prefix/phases", helper.DocEntryPointPrefix())
}

func TestHelperEmptyDocEntryPointPrefix(t *testing.T) {
	cfg := testConfig(t)
	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)
	assert.Equal(t, "", helper.DocEntryPointPrefix())
}

func TestHelperWorkdir(t *testing.T) {
	helper, err := phase.NewHelper(testConfig(t))
	require.NoError(t, err)
	require.NotNil(t, helper)
	workDir, err := helper.WorkDir()
	assert.NoError(t, err)
	assert.Greater(t, len(workDir), 0)
}

func TestHelperInventory(t *testing.T) {
	helper, err := phase.NewHelper(testConfig(t))
	require.NoError(t, err)
	require.NotNil(t, helper)
	inv, err := helper.Inventory()
	assert.NoError(t, err)
	assert.NotNil(t, inv)
}

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	confString := `apiVersion: airshipit.org/v1alpha1
contexts:
  dummy_cluster:
    contextKubeconf: dummy_cluster
    manifest: dummy_manifest
currentContext: dummy_cluster
kind: Config
manifests:
  dummy_manifest:
    phaseRepositoryName: primary
    targetPath: testdata
    metadataPath: valid_site/metadata.yaml
    repositories:
      primary:
        url: "empty/filename/"
`
	conf := &config.Config{}
	err := yaml.Unmarshal([]byte(confString), conf)
	require.NoError(t, err)
	return conf
}

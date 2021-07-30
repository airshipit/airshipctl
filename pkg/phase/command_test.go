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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

const (
	testFactoryErr        = "test config error"
	testNewHelperErr      = "missing configuration"
	testNoBundlePath      = "no such file or directory"
	defaultCurrentContext = "context"
	testTargetPath        = "testdata"
	testMetadataPath      = "metadata.yaml"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name        string
		errContains string
		runFlags    ifc.RunOptions
		factory     config.Factory
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, fmt.Errorf(testFactoryErr)
			},
			errContains: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			errContains: testNewHelperErr,
		},
		{
			name: "Error phase by id",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "broken_metadata.yaml",
						TargetPath:          testTargetPath,
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			errContains: testNoBundlePath,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			command := phase.RunCommand{
				Options: tt.runFlags,
				Factory: tt.factory,
			}
			err := command.RunE()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListCommand(t *testing.T) {
	outputString1 := "NAMESPACE   RESOURCE                                  CLUSTER " +
		"NAME          EXECUTOR              DOC ENTRYPOINT                     " +
		"                                                                 "
	outputString2 := "            Phase/phase                               ephemeral" +
		"-cluster     KubernetesApply       ephemeral/phase                      " +
		"                                                               "
	yamlOutput := `---
- apiVersion: airshipit.org/v1alpha1
  config:
    documentEntryPoint: ephemeral/phase
    executorRef:
      apiVersion: airshipit.org/v1alpha1
      kind: KubernetesApply
      name: kubernetes-apply
    validation: {}
  kind: Phase
  metadata:
    clusterName: ephemeral-cluster
    creationTimestamp: null
    name: phase
...
`
	tests := []struct {
		name            string
		errContains     string
		runFlags        phase.RunFlags
		expectedOut     [][]byte
		expectedYamlOut string
		factory         config.Factory
		PlanID          ifc.ID
		PhaseID         ifc.ID
		OutputFormat    string
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, fmt.Errorf(testFactoryErr)
			},
			errContains:  testFactoryErr,
			expectedOut:  [][]byte{{}},
			OutputFormat: "table",
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			errContains:  testNewHelperErr,
			expectedOut:  [][]byte{{}},
			OutputFormat: "table",
		},
		{
			name: "List phases",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        testMetadataPath,
						TargetPath:          testTargetPath,
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			expectedOut: [][]byte{
				[]byte(outputString1),
				[]byte(outputString2),
				{},
			},
			OutputFormat: "table",
		},
		{
			name: "List phases of a plan",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				manifest := conf.Manifests[config.AirshipDefaultManifest]
				manifest.TargetPath = testTargetPath
				manifest.MetadataPath = testMetadataPath
				manifest.Repositories[config.DefaultTestPhaseRepo].URLString = ""
				return conf, nil
			},
			PlanID: ifc.ID{Name: "phasePlan"},
			expectedOut: [][]byte{
				[]byte(outputString1),
				[]byte(outputString2),
				{},
			},
			OutputFormat: "table",
		},
		{
			name: "List phases yaml format",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				manifest := conf.Manifests[config.AirshipDefaultManifest]
				manifest.TargetPath = testTargetPath
				manifest.MetadataPath = testMetadataPath
				manifest.Repositories[config.DefaultTestPhaseRepo].URLString = ""
				return conf, nil
			},
			OutputFormat:    "yaml",
			expectedYamlOut: yamlOutput,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			command := phase.ListCommand{
				Factory:      tt.factory,
				Writer:       buffer,
				PlanID:       tt.PlanID,
				OutputFormat: tt.OutputFormat,
			}
			err := command.RunE()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
			out, err := ioutil.ReadAll(buffer)
			require.NoError(t, err)
			if tt.OutputFormat == "yaml" {
				assert.Equal(t, tt.expectedYamlOut, string(out))
			} else {
				b := bytes.Split(out, []byte("\n"))
				assert.Equal(t, tt.expectedOut, b)
			}
		})
	}
}

func TestTreeCommand(t *testing.T) {
	tests := []struct {
		name        string
		errContains string
		factory     config.Factory
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, fmt.Errorf(testFactoryErr)
			},
			errContains: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			errContains: testNewHelperErr,
		},
		{
			name: "Error phase by id",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "broken_metadata.yaml",
						TargetPath:          testTargetPath,
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			errContains: testNoBundlePath,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			command := phase.TreeCommand{
				Factory: tt.factory,
			}
			err := command.RunE()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanListCommand(t *testing.T) {
	yamlOutput := `---
- apiVersion: airshipit.org/v1alpha1
  description: Default phase plan
  kind: PhasePlan
  metadata:
    creationTimestamp: null
    name: phasePlan
  phases:
  - name: phase
  validation: {}
...
`
	testErr := fmt.Errorf(testFactoryErr)
	testCases := []struct {
		name         string
		factory      config.Factory
		expectedOut  [][]byte
		expectedErr  string
		Format       string
		expectedYaml string
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, testErr
			},
			expectedErr: testFactoryErr,
			expectedOut: [][]byte{{}},
			Format:      "table",
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			expectedErr: testNewHelperErr,
			Format:      "table",
			expectedOut: [][]byte{{}},
		},

		{
			name: "List phases",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				manifest := conf.Manifests[config.AirshipDefaultManifest]
				manifest.TargetPath = testTargetPath
				manifest.MetadataPath = testMetadataPath
				manifest.Repositories[config.DefaultTestPhaseRepo].URLString = ""
				return conf, nil
			},
			expectedOut: [][]byte{
				[]byte("NAMESPACE   RESOURCE                                  DESCRIPTION                                        " +
					"                                                                                                        " +
					"                                             "),
				[]byte("            PhasePlan/phasePlan                       Default phase plan" +
					"                                                                            " +
					"                                                                            " +
					"                              "),
				{},
			},
			Format: "table",
		},
		{
			name: "Valid yaml input format",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				manifest := conf.Manifests[config.AirshipDefaultManifest]
				manifest.TargetPath = testTargetPath
				manifest.MetadataPath = "metadata.yaml"
				manifest.Repositories[config.DefaultTestPhaseRepo].URLString = ""
				return conf, nil
			},
			Format:       "yaml",
			expectedYaml: yamlOutput,
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			cmd := phase.PlanListCommand{
				Factory: tt.factory,
				Writer:  buf,
				Options: phase.PlanListFlags{FormatType: tt.Format},
			}
			err := cmd.RunE()
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			out, err := ioutil.ReadAll(buf)
			fmt.Print(string(out))
			require.NoError(t, err)
			if tt.Format == "yaml" {
				assert.Equal(t, tt.expectedYaml, string(out))
			} else {
				assert.Equal(t, tt.expectedOut, bytes.Split(out, []byte("\n")))
			}
		})
	}
}

func TestPlanRunCommand(t *testing.T) {
	log.Init(true, os.Stdout)
	testErr := fmt.Errorf(testFactoryErr)
	testCases := []struct {
		name        string
		factory     config.Factory
		expectedErr string
		planID      ifc.ID
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, testErr
			},
			expectedErr: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			expectedErr: "missing configuration: context with name 'does not exist'",
		},
		{
			name: "Error plan by id",
			planID: ifc.ID{
				Name: "doesn't exist",
			},
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        testMetadataPath,
						TargetPath:          testTargetPath,
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			expectedErr: `found no documents`,
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.PlanRunCommand{
				Options: ifc.RunOptions{
					DryRun: true,
				},
				Factory: tt.factory,
				PlanID:  tt.planID,
			}
			err := cmd.RunE()
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClusterListCommand_RunE(t *testing.T) {
	testErr := fmt.Errorf(testFactoryErr)
	testCases := []struct {
		name        string
		factory     config.Factory
		expectedErr string
		Format      string
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, testErr
			},
			expectedErr: testFactoryErr,
			Format:      "name",
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			expectedErr: "missing configuration: context with name 'does not exist'",
			Format:      "name",
		},
		{
			name:   "No error",
			Format: "name",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "metadata.yaml",
						TargetPath:          "testdata",
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			expectedErr: "",
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.ClusterListCommand{
				Factory: tt.factory,
				Format:  tt.Format,
				Writer:  bytes.NewBuffer(nil),
			}
			err := cmd.RunE()
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		errContains string
		flags       phase.ValidateFlags
		factory     config.Factory
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, fmt.Errorf(testFactoryErr)
			},
			errContains: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			errContains: testNewHelperErr,
		},
		{
			name: "Error phase by id",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "broken_metadata.yaml",
						TargetPath:          "testdata",
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			errContains: testNoBundlePath,
		},
		{
			name: "success",
			// flags: phase.ValidateFlags{PhaseID: }
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "metadata.yaml",
						TargetPath:          "testdata",
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			errContains: `document filtered by selector [Group="airshipit.org", Version="v1alpha1", ` +
				`Kind="GenericContainer", Name="document-validation"] found no documents`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			command := phase.ValidateCommand{
				Options: tt.flags,
				Factory: tt.factory,
			}
			err := command.RunE()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStatusCommand(t *testing.T) {
	tests := []struct {
		name        string
		errContains string
		statusFlags phase.StatusFlags
		factory     config.Factory
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, fmt.Errorf(testFactoryErr)
			},
			errContains: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			errContains: testNewHelperErr,
		},
		{
			name: "Error phase by id",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "broken_metadata.yaml",
						TargetPath:          "testdata",
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = "context"
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			errContains: testNoBundlePath,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			command := phase.StatusCommand{
				Options: tt.statusFlags,
				Factory: tt.factory,
			}
			err := command.RunE()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPlanValidateCommand(t *testing.T) {
	testErr := fmt.Errorf(testFactoryErr)
	testCases := []struct {
		name        string
		factory     config.Factory
		expectedErr string
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, testErr
			},
			expectedErr: testFactoryErr,
		},
		{
			name: "Error new helper",
			factory: func() (*config.Config, error) {
				return &config.Config{
					CurrentContext: "does not exist",
					Contexts:       make(map[string]*config.Context),
				}, nil
			},
			expectedErr: "missing configuration: context with name 'does not exist'",
		},
		{
			name: "Error plan by id",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				conf.Manifests = map[string]*config.Manifest{
					"manifest": {
						MetadataPath:        "metadata.yaml",
						TargetPath:          "testdata",
						PhaseRepositoryName: config.DefaultTestPhaseRepo,
						Repositories: map[string]*config.Repository{
							config.DefaultTestPhaseRepo: {
								URLString: "",
							},
						},
					},
				}
				conf.CurrentContext = defaultCurrentContext
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			expectedErr: `found no documents`,
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.PlanValidateCommand{
				Options: phase.PlanValidateFlags{PlanID: ifc.ID{Name: "invalid"}},
				Factory: tt.factory,
			}
			err := cmd.RunE()
			if tt.expectedErr != "" {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

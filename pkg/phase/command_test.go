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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	testFactoryErr   = "test config error"
	testNewHelperErr = "Missing configuration"
	testNoBundlePath = "no such file or directory"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name        string
		errContains string
		runFlags    phase.RunFlags
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
	tests := []struct {
		name        string
		errContains string
		runFlags    phase.RunFlags
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
			command := phase.ListCommand{
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
	testErr := fmt.Errorf(testFactoryErr)
	testCases := []struct {
		name        string
		factory     config.Factory
		expectedOut [][]byte
		expectedErr string
	}{
		{
			name: "Error config factory",
			factory: func() (*config.Config, error) {
				return nil, testErr
			},
			expectedErr: testFactoryErr,
			expectedOut: [][]byte{{}},
		},
		{
			name: "List phases",
			factory: func() (*config.Config, error) {
				conf := config.NewConfig()
				manifest := conf.Manifests[config.AirshipDefaultManifest]
				manifest.TargetPath = "testdata"
				manifest.MetadataPath = "metadata.yaml"
				manifest.Repositories[config.DefaultTestPhaseRepo].URLString = ""
				return conf, nil
			},
			expectedOut: [][]byte{
				[]byte("NAMESPACE   RESOURCE                                  DESCRIPTION                             "),
				[]byte("            PhasePlan/phasePlan                       Default phase plan                      "),
				{},
			},
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			cmd := phase.PlanListCommand{
				Factory: tt.factory,
				Writer:  buf,
			}
			err := cmd.RunE()
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			out, err := ioutil.ReadAll(buf)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOut, bytes.Split(out, []byte("\n")))
		})
	}
}

func TestPlanRunCommand(t *testing.T) {
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
			expectedErr: "Missing configuration: Context with name 'does not exist'",
		},
		{
			name: "Error phase by id",
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
				conf.CurrentContext = "context"
				conf.Contexts = map[string]*config.Context{
					"context": {
						Manifest: "manifest",
					},
				}
				return conf, nil
			},
			expectedErr: `Error events received on channel, errors are:
[document filtered by selector [Group="airshipit.org", Version="v1alpha1", Kind="KubeConfig"] found no documents]`,
		},
	}
	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			cmd := phase.PlanRunCommand{
				Options: phase.PlanRunFlags{
					GenericRunFlags: phase.GenericRunFlags{
						DryRun: true,
					},
				},
				Factory: tt.factory,
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

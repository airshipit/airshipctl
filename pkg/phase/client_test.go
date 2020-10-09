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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/events"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/phase/ifc"
)

func TestClientPhaseExecutor(t *testing.T) {
	tests := []struct {
		name             string
		errContains      string
		expectedExecutor ifc.Executor
		phaseID          ifc.ID
		configFunc       func(t *testing.T) *config.Config
		registryFunc     phase.ExecutorRegistry
	}{
		{
			name:             "Success fake executor",
			expectedExecutor: fakeExecutor{},
			configFunc:       testConfig,
			phaseID:          ifc.ID{Name: "capi_init"},
			registryFunc:     fakeRegistry,
		},
		{
			name:             "Error executor doc doesn't exist",
			expectedExecutor: fakeExecutor{},
			configFunc:       testConfig,
			phaseID:          ifc.ID{Name: "some_phase"},
			registryFunc:     fakeRegistry,
			errContains:      "found no documents",
		},
		{
			name:             "Error executor doc not registered",
			expectedExecutor: fakeExecutor{},
			configFunc:       testConfig,
			phaseID:          ifc.ID{Name: "capi_init"},
			registryFunc: func() map[schema.GroupVersionKind]ifc.ExecutorFactory {
				return make(map[schema.GroupVersionKind]ifc.ExecutorFactory)
			},
			errContains: "executor identified by 'airshipit.org/v1alpha1, Kind=Clusterctl' is not found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			conf := tt.configFunc(t)
			helper, err := phase.NewHelper(conf)
			require.NoError(t, err)
			require.NotNil(t, helper)
			client := phase.NewClient(helper, phase.InjectRegistry(tt.registryFunc))
			require.NotNil(t, client)
			p, err := client.PhaseByID(tt.phaseID)
			require.NotNil(t, client)
			executor, err := p.Executor()
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Nil(t, executor)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
				require.NotNil(t, executor)
				assertEqualExecutor(t, tt.expectedExecutor, executor)
			}
		})
	}
}

func TestPhaseRun(t *testing.T) {
	tests := []struct {
		name         string
		errContains  string
		phaseID      ifc.ID
		configFunc   func(t *testing.T) *config.Config
		registryFunc phase.ExecutorRegistry
	}{
		{
			name:         "Success fake executor",
			configFunc:   testConfig,
			phaseID:      ifc.ID{Name: "capi_init"},
			registryFunc: fakeRegistry,
		},
		{
			name:         "Error executor doc doesn't exist",
			configFunc:   testConfig,
			phaseID:      ifc.ID{Name: "some_phase"},
			registryFunc: fakeRegistry,
			errContains:  "found no documents",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			conf := tt.configFunc(t)
			helper, err := phase.NewHelper(conf)
			require.NoError(t, err)
			require.NotNil(t, helper)
			client := phase.NewClient(helper, phase.InjectRegistry(tt.registryFunc))
			require.NotNil(t, client)
			p, err := client.PhaseByID(tt.phaseID)
			require.NotNil(t, client)
			err = p.Run(ifc.RunOptions{DryRun: true})
			if tt.errContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TODO develop tests, when we add phase object validation
func TestClientByAPIObj(t *testing.T) {
	helper, err := phase.NewHelper(testConfig(t))
	require.NoError(t, err)
	require.NotNil(t, helper)
	client := phase.NewClient(helper)
	require.NotNil(t, client)
	p, err := client.PhaseByAPIObj(&v1alpha1.Phase{})
	require.NoError(t, err)
	require.NotNil(t, p)
}

// assertEqualExecutor allows to compare executor interfaces
// check if we expect nil, and if so actual interface must be nil also otherwise compare types
func assertEqualExecutor(t *testing.T, expected, actual ifc.Executor) {
	if expected == nil {
		assert.Nil(t, actual)
		return
	}
	assert.IsType(t, expected, actual)
}

func fakeRegistry() map[schema.GroupVersionKind]ifc.ExecutorFactory {
	gvk := schema.GroupVersionKind{
		Group:   "airshipit.org",
		Version: "v1alpha1",
		Kind:    "Clusterctl",
	}
	return map[schema.GroupVersionKind]ifc.ExecutorFactory{
		gvk: fakeExecFactory,
	}
}

func TestDocumentRoot(t *testing.T) {
	tests := []struct {
		name         string
		expectedRoot string
		phaseID      ifc.ID
		expectedErr  error
		metadataPath string
	}{
		{
			name:         "Success entrypoint exists",
			expectedRoot: "testdata/valid_site/phases",
			phaseID:      ifc.ID{Name: "capi_init"},
		},
		{
			name:        "Error entrypoint does not exists",
			phaseID:     ifc.ID{Name: "some_phase"},
			expectedErr: phase.ErrDocumentEntrypointNotDefined{PhaseName: "some_phase"},
		},
		{
			name:         "Success entrypoint with doc prefix path",
			expectedRoot: "testdata/valid_site_with_doc_prefix/phases/entrypoint",
			phaseID:      ifc.ID{Name: "sample"},
			metadataPath: "valid_site_with_doc_prefix/metadata.yaml",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig(t)
			if tt.metadataPath != "" {
				cfg.Manifests["dummy_manifest"].MetadataPath = tt.metadataPath
			}
			helper, err := phase.NewHelper(cfg)
			require.NoError(t, err)
			require.NotNil(t, helper)

			c := phase.NewClient(helper)
			p, err := c.PhaseByID(ifc.ID{Name: tt.phaseID.Name, Namespace: tt.phaseID.Namespace})
			require.NoError(t, err)
			require.NotNil(t, p)

			root, err := p.DocumentRoot()
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedRoot, root)
		})
	}
}

func TestBundleFactoryExecutor(t *testing.T) {
	cfg := testConfig(t)

	helper, err := phase.NewHelper(cfg)
	require.NoError(t, err)
	require.NotNil(t, helper)

	fakeRegistry := func() map[schema.GroupVersionKind]ifc.ExecutorFactory {
		validBundleFactory := schema.GroupVersionKind{
			Group:   "airshipit.org",
			Version: "v1alpha1",
			Kind:    "Clusterctl",
		}
		invalidBundleFactory := schema.GroupVersionKind{
			Group:   "airshipit.org",
			Version: "v1alpha1",
			Kind:    "SomeExecutor",
		}
		bundleCheckFunc := func(config ifc.ExecutorConfig) (ifc.Executor, error) {
			_, bundleErr := config.BundleFactory()
			return nil, bundleErr
		}
		return map[schema.GroupVersionKind]ifc.ExecutorFactory{
			// validBundleFactory has DocumentEntryPoint defined
			validBundleFactory: bundleCheckFunc,
			// invalidBundleFactory does not have DocumentEntryPoint defined, error should be returned
			invalidBundleFactory: bundleCheckFunc,
		}
	}
	c := phase.NewClient(helper, phase.InjectRegistry(fakeRegistry))
	p, err := c.PhaseByID(ifc.ID{Name: "capi_init"})
	require.NoError(t, err)
	_, err = p.Executor()
	assert.NoError(t, err)
	p, err = c.PhaseByID(ifc.ID{Name: "no_entry_point"})
	require.NoError(t, err)
	_, err = p.Executor()
	assert.Error(t, err)
	assert.Equal(t, phase.ErrDocumentEntrypointNotDefined{PhaseName: "no_entry_point"}, err)
}

func fakeExecFactory(config ifc.ExecutorConfig) (ifc.Executor, error) {
	return fakeExecutor{}, nil
}

var _ ifc.Executor = fakeExecutor{}

type fakeExecutor struct {
}

func (e fakeExecutor) Render(_ io.Writer, _ ifc.RenderOptions) error {
	return nil
}

func (e fakeExecutor) Run(ch chan events.Event, _ ifc.RunOptions) {
	defer close(ch)
}

func (e fakeExecutor) Validate() error {
	return nil
}

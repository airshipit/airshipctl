/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kustfs "sigs.k8s.io/kustomize/api/filesys"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
	testfs "opendev.org/airship/airshipctl/testutil/fs"
)

func prepareConfig() func() (*config.Config, error) {
	return func() (*config.Config, error) {
		cfg := testutil.DummyConfig()
		cfg.SetLoadedConfigPath("test")
		cfg.SetFs(testfs.MockFileSystem{
			FileSystem: kustfs.MakeFsInMemory(),
			MockChmod: func(s string, mode os.FileMode) error {
				return nil
			},
			MockDir: func(s string) string {
				return "."
			},
		})
		return cfg, nil
	}
}

func TestRunSetContext(t *testing.T) {
	ioBuffer := bytes.NewBuffer(nil)
	tests := []struct {
		name        string
		options     config.RunSetContextOptions
		ctxopts     []config.ContextOption
		expectedOut string
		err         error
	}{
		{
			name: "create new context",
			options: config.RunSetContextOptions{
				CfgFactory: prepareConfig(),
				CtxName:    "new_context",
				Current:    false,
				Writer:     ioBuffer,
			},
			ctxopts: []config.ContextOption{config.SetContextManifest("dummy_manifest"),
				config.SetContextManagementConfig("dummy_management_config")},
			err:         nil,
			expectedOut: "context with name new_context created\n",
		},
		{
			name: "modify current context",
			options: config.RunSetContextOptions{
				CfgFactory: prepareConfig(),
				CtxName:    "",
				Current:    true,
				Writer:     ioBuffer,
			},
			ctxopts:     []config.ContextOption{config.SetContextManagementConfig("")},
			err:         nil,
			expectedOut: "context with name dummy_context modified\n",
		},
		{
			name: "bad config",
			options: config.RunSetContextOptions{
				CfgFactory: func() (*config.Config, error) {
					return nil, config.ErrMissingConfig{What: "bad config"}
				},
			},
			err: config.ErrMissingConfig{What: "bad config"},
		},
		{
			name: "no context name provided",
			options: config.RunSetContextOptions{
				CfgFactory: prepareConfig(),
				CtxName:    "",
				Current:    false,
			},
			err: config.ErrEmptyContextName{},
		},
		{
			name: "setup invalid manifest",
			options: config.RunSetContextOptions{
				CfgFactory: prepareConfig(),
				CtxName:    "",
				Current:    true,
				Writer:     ioBuffer,
			},
			ctxopts: []config.ContextOption{config.SetContextManifest("invalid_manifest")},
			err:     config.ErrMissingConfig{What: "Current Context (dummy_context) does not identify a defined Manifest"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ioBuffer.Reset()
			err := tt.options.RunSetContext(tt.ctxopts...)
			if tt.err != nil {
				require.Error(t, err)
				require.Equal(t, tt.err, err)
			} else {
				require.NoError(t, err)
			}
			if tt.expectedOut != "" {
				require.Equal(t, tt.expectedOut, ioBuffer.String())
			}
		})
	}
}

func TestRunUseContext(t *testing.T) {
	t.Run("testUseContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		err := config.RunUseContext("dummy_context", conf)
		assert.Nil(t, err)
	})

	t.Run("testUseContextDoesNotExist", func(t *testing.T) {
		conf := config.NewConfig()
		err := config.RunUseContext("foo", conf)
		assert.Error(t, err)
	})
}

func TestRunSetManifest(t *testing.T) {
	t.Run("testAddManifest", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyManifestOptions := testutil.DummyManifestOptions()
		dummyManifestOptions.Name = "test_manifest"

		modified, err := config.RunSetManifest(dummyManifestOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
	})

	t.Run("testModifyManifest", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyManifestOptions := testutil.DummyManifestOptions()
		dummyManifestOptions.TargetPath = "/tmp/default"

		modified, err := config.RunSetManifest(dummyManifestOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "/tmp/default", conf.Manifests["dummy_manifest"].TargetPath)
	})
}

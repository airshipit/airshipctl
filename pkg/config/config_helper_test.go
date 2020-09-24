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
	"testing"

	"github.com/stretchr/testify/assert"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

func TestRunSetContext(t *testing.T) {
	t.Run("testAddContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()
		dummyContextOptions.Name = "second_context"

		modified, err := config.RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.False(t, modified)
		assert.Contains(t, conf.Contexts, "second_context")
	})

	t.Run("testModifyContext", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyContextOptions := testutil.DummyContextOptions()

		modified, err := config.RunSetContext(dummyContextOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
	})
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

func TestRunSetEncryptionConfigLocalFile(t *testing.T) {
	t.Run("testAddEncryptionConfig", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyEncryptionConfig := testutil.DummyEncryptionConfigOptions()
		dummyEncryptionConfig.Name = "test_encryption_config"

		modified, err := config.RunSetEncryptionConfig(dummyEncryptionConfig, conf, false)
		assert.Error(t, err)
		assert.False(t, modified)
	})

	t.Run("testModifyEncryptionConfig", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyEncryptionConfigOptions := &config.EncryptionConfigOptions{
			Name:              "testModifyEncryptionConfig",
			EncryptionKeyPath: "testdata/ca.crt",
			DecryptionKeyPath: "testdata/test-key.pem",
		}

		modified, err := config.RunSetEncryptionConfig(dummyEncryptionConfigOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "testdata/ca.crt", conf.EncryptionConfigs["testModifyEncryptionConfig"].EncryptionKeyPath)
		assert.Equal(t, "testdata/test-key.pem", conf.EncryptionConfigs["testModifyEncryptionConfig"].DecryptionKeyPath)
	})
}

func TestRunSetEncryptionConfigAPIBackend(t *testing.T) {
	t.Run("testAddEncryptionConfig", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyEncryptionConfig := testutil.DummyEncryptionConfigOptions()
		dummyEncryptionConfig.Name = "test_encryption_config"

		modified, err := config.RunSetEncryptionConfig(dummyEncryptionConfig, conf, false)
		assert.Error(t, err)
		assert.False(t, modified)
	})

	t.Run("testModifyEncryptionConfig", func(t *testing.T) {
		conf := testutil.DummyConfig()
		dummyEncryptionConfigOptions := &config.EncryptionConfigOptions{
			Name:               "testModifyEncryptionConfig",
			KeySecretName:      "dummySecret",
			KeySecretNamespace: "dummyNamespace",
			EncryptionKeyPath:  "",
			DecryptionKeyPath:  "",
		}

		modified, err := config.RunSetEncryptionConfig(dummyEncryptionConfigOptions, conf, false)
		assert.NoError(t, err)
		assert.True(t, modified)
		assert.Equal(t, "dummySecret", conf.EncryptionConfigs["testModifyEncryptionConfig"].KeySecretName)
		assert.Equal(t, "dummyNamespace", conf.EncryptionConfigs["testModifyEncryptionConfig"].KeySecretNamespace)
	})
}

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

package testutil

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// types cloned directory from pkg/config/types to prevent circular import

// DummyConfig used by tests, to initialize min set of data
func DummyConfig() *config.Config {
	conf := &config.Config{
		Kind:       config.AirshipConfigKind,
		APIVersion: config.AirshipConfigAPIVersion,
		Permissions: config.Permissions{
			DirectoryPermission: config.AirshipDefaultDirectoryPermission,
			FilePermission:      config.AirshipDefaultFilePermission,
		},
		Contexts: map[string]*config.Context{
			"dummy_context": DummyContext(),
		},
		Manifests: map[string]*config.Manifest{
			"dummy_manifest": DummyManifest(),
		},
		ManagementConfiguration: map[string]*config.ManagementConfiguration{
			"dummy_management_config": DummyManagementConfiguration(),
		},
		EncryptionConfigs: map[string]*config.EncryptionConfig{
			"dummy_encryption_config": DummyEncryptionConfig(),
		},
		CurrentContext: "dummy_context",
	}
	return conf
}

// DummyContext creates a Context config object for unit testing
func DummyContext() *config.Context {
	c := config.NewContext()
	c.NameInKubeconf = "dummy_cluster_ephemeral"
	c.Manifest = "dummy_manifest"
	c.EncryptionConfig = "dummy_encryption_config"
	c.ManagementConfiguration = "dummy_management_config"
	return c
}

// DummyManifest creates a Manifest config object for unit testing
func DummyManifest() *config.Manifest {
	m := config.NewManifest()
	// Repositories is the map of repository addressable by a name
	m.Repositories = map[string]*config.Repository{"primary": DummyRepository()}
	m.PrimaryRepositoryName = "primary"
	m.TargetPath = "/var/tmp/"
	m.SubPath = "manifests/site/test-site"
	return m
}

// DummyRepository creates a Repository config object for unit testing
func DummyRepository() *config.Repository {
	return &config.Repository{
		URLString: "http://dummy.url.com/manifests.git",
		CheckoutOptions: &config.RepoCheckout{
			Tag:           "v1.0.1",
			ForceCheckout: false,
		},
		Auth: &config.RepoAuth{
			Type:    "ssh-key",
			KeyPath: "testdata/test-key.pem",
		},
	}
}

// DummyRepoAuth creates a RepoAuth config object for unit testing
func DummyRepoAuth() *config.RepoAuth {
	return &config.RepoAuth{
		Type:    "ssh-key",
		KeyPath: "testdata/test-key.pem",
	}
}

// DummyRepoCheckout creates a RepoCheckout config object
// for unit testing
func DummyRepoCheckout() *config.RepoCheckout {
	return &config.RepoCheckout{
		Tag:           "v1.0.1",
		ForceCheckout: false,
	}
}

// InitConfig creates a Config object meant for testing.
//
// The returned config object will be associated with real files stored in a
// directory in the user's temporary file storage
// This directory can be cleaned up by calling the returned "cleanup" function
func InitConfig(t *testing.T) (conf *config.Config, cleanup func(*testing.T)) {
	t.Helper()
	testDir, cleanup := TempDir(t, "airship-test")

	configPath := filepath.Join(testDir, "config")
	err := ioutil.WriteFile(configPath, []byte(testConfigYAML), 0600)
	require.NoError(t, err)

	conf = config.NewConfig()

	cfg, err := config.CreateFactory(&configPath)()
	require.NoError(t, err)

	return cfg, cleanup
}

// DummyContextOptions creates ContextOptions config object
// for unit testing
func DummyContextOptions() *config.ContextOptions {
	co := &config.ContextOptions{}
	co.Name = "dummy_context"
	co.Manifest = "dummy_manifest"
	co.CurrentContext = false
	co.EncryptionConfig = "dummy_encryption_config"
	return co
}

// DummyEncryptionConfig creates EncryptionConfigOptions object
// for unit testing
func DummyEncryptionConfig() *config.EncryptionConfig {
	return &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			EncryptionKeyPath: "/tmp/encryption.key",
			DecryptionKeyPath: "/tmp/decryption.pub",
		},
	}
}

// DummyEncryptionConfigOptions creates ManifestOptions config object
// for unit testing
func DummyEncryptionConfigOptions() *config.EncryptionConfigOptions {
	return &config.EncryptionConfigOptions{
		Name:              "dummy_encryption_config",
		EncryptionKeyPath: "/tmp/encryption.key",
		DecryptionKeyPath: "/tmp/decryption.pub",
	}
}

// DummyManagementConfiguration creates a management configuration for unit testing
func DummyManagementConfiguration() *config.ManagementConfiguration {
	return &config.ManagementConfiguration{
		Type:     redfish.ClientType,
		Insecure: true,
		UseProxy: false,
	}
}

// DummyManifestOptions creates ManifestOptions config object
// for unit testing
func DummyManifestOptions() *config.ManifestOptions {
	return &config.ManifestOptions{
		Name:       "dummy_manifest",
		SubPath:    "manifests/dummy_site",
		TargetPath: "/tmp/dummy_site",
		IsPrimary:  true,
		RepoName:   "dummy_repo",
		URL:        "https://github.com/treasuremap/dummy_site",
		Branch:     "master",
		Force:      true,
	}
}

const (
	testConfigYAML = `apiVersion: airshipit.org/v1alpha1
contexts:
  def_ephemeral:
    contextKubeconf: def_ephemeral
    manifest: dummy_manifest
  def_target:
    contextKubeconf: def_target
  onlyink:
    contextKubeconf: onlyinkubeconf_target
encryptionConfigs: {}
currentContext: def_ephemeral
kind: Config
manifests:
  dummy_manifest: {}`
)

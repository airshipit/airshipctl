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

package secret

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	qualifiedEncryptedSecret = "testdata/secrets/decryption/qualified-encrypted-secret.yaml"
	encryptionKey            = "testdata/encryption.pub"
	decryptionKey            = "testdata/decryption.key"

	currentContext       = "def_ephemeral"
	testEncryptionConfig = "test"
	testManifest         = "test"
)

func TestDecrypt(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs[testEncryptionConfig] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: decryptionKey,
			EncryptionKeyPath: encryptionKey,
		},
	}
	ctx, err := cfg.GetContext(currentContext)
	ctx.EncryptionConfig = testEncryptionConfig
	require.NoError(t, err)
	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt-invalid-public-key")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = decrypt(cfg, fake.NewClient(), qualifiedEncryptedSecret, tmpFile.Name())
	assert.NoError(t, err)
}

func TestDecryptWithContextPath(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs[testEncryptionConfig] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: decryptionKey,
			EncryptionKeyPath: encryptionKey,
		},
	}
	manifest := &config.Manifest{
		TargetPath:          "testdata/secrets/decryption/",
		MetadataPath:        "metadata.yaml",
		Repositories:        map[string]*config.Repository{"primary": testutil.DummyRepository()},
		PhaseRepositoryName: "primary",
	}
	if cfg.Manifests == nil {
		cfg.Manifests = make(map[string]*config.Manifest)
	}
	cfg.Manifests[testManifest] = manifest
	ctx, err := cfg.GetCurrentContext()
	ctx.EncryptionConfig = testEncryptionConfig
	ctx.Manifest = testManifest
	require.NoError(t, err)
	err = decrypt(cfg, fake.NewClient(), "", "")
	assert.NoError(t, err)
}

func TestDecryptInvalidPrivateKey(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs[testEncryptionConfig] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: "dummy",
			EncryptionKeyPath: "testdata/encryption.pub",
		},
	}
	ctx, err := cfg.GetContext(currentContext)
	ctx.EncryptionConfig = testEncryptionConfig
	require.NoError(t, err)
	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt-invalid-public-key")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = decrypt(cfg, fake.NewClient(), qualifiedEncryptedSecret, tmpFile.Name())
	assert.Error(t, err)
}

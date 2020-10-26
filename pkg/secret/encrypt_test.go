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
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/k8s/client/fake"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	qualifiedDecryptedFile = "testdata/secrets/encryption/qualified-decrypted-secret.yaml"

	keyID = "681E3A89EB1DAFD36EB883120A73BB48E26694D8"
)

func TestEncrypt(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs["test"] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: "testdata/decryption.key",
			EncryptionKeyPath: "testdata/encryption.pub",
		},
	}
	ctx, err := cfg.GetContext(currentContext)
	require.NoError(t, err)
	ctx.EncryptionConfig = testEncryptionConfig

	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = encrypt(cfg, fake.NewClient(), qualifiedDecryptedFile, tmpFile.Name())
	assert.NoError(t, err)
}

func TestEncryptWithContextPath(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs[testEncryptionConfig] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: "testdata/decryption.key",
			EncryptionKeyPath: "testdata/encryption.pub",
		},
	}
	manifest := &config.Manifest{
		TargetPath:          "testdata/secrets/encryption/",
		MetadataPath:        "metadata.yaml",
		Repositories:        map[string]*config.Repository{"primary": testutil.DummyRepository()},
		PhaseRepositoryName: "primary",
	}
	if cfg.Manifests == nil {
		cfg.Manifests = make(map[string]*config.Manifest)
	}
	cfg.Manifests[testManifest] = manifest
	ctx, err := cfg.GetCurrentContext()
	require.NoError(t, err)
	ctx.EncryptionConfig = testEncryptionConfig
	ctx.Manifest = testManifest
	dir, err := ioutil.TempDir("/tmp/", "encrypt-context-path")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = encrypt(cfg, fake.NewClient(), "", dir)
	assert.NoError(t, err)
}

func TestEncryptInvalidPublicKey(t *testing.T) {
	defer deleteGpgKeys()
	cfg, _ := testutil.InitConfig(t)
	cfg.CurrentContext = currentContext
	cfg.EncryptionConfigs[testEncryptionConfig] = &config.EncryptionConfig{
		EncryptionKeyFileSource: config.EncryptionKeyFileSource{
			DecryptionKeyPath: "testdata/decryption.key",
			EncryptionKeyPath: "testdata/decryption.key",
		},
	}
	ctx, err := cfg.GetContext(currentContext)
	require.NoError(t, err)
	ctx.EncryptionConfig = testEncryptionConfig

	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt-invalid-public-key")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = encrypt(cfg, fake.NewClient(), qualifiedDecryptedFile, tmpFile.Name())
	assert.Error(t, err)
}

func deleteGpgKeys() {
	gpgCmd := exec.Command("gpg", "--delete-secret-keys", "--batch", "--yes", keyID)
	if err := gpgCmd.Run(); err != nil {
		// best effort to delete the secret keys
		return
	}
	gpgCmd = exec.Command("gpg", "--delete-keys", "--batch", "--yes", keyID)
	if err := gpgCmd.Run(); err != nil {
		// best effort to delete the secret keys
		return
	}
	return
}

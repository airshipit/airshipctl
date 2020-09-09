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

package sops_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/secret/sops"
)

const (
	qualifiedEncryptedFile          = "testdata/secrets/qualified-encrypted-secret.yaml"
	qualifiedDecryptedFile          = "testdata/secrets/qualified-decrypted-secret.yaml"
	qualifiedDecryptedFileWithRegex = "testdata/secrets/qualified-decrypted-secret-with-regex.yaml"
	invalidYamlDecryptedFile        = "testdata/secrets/qualified-decrypted-invalid-yaml.yaml"
	missingMetadataEncryptedFile    = "testdata/secrets/qualified-encrypted-secret-missing-metadata.yaml"

	keyID = "681E3A89EB1DAFD36EB883120A73BB48E26694D8"
)

func TestEncrypt(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = sopsClient.Encrypt(qualifiedDecryptedFile, tmpFile.Name())
	assert.NoError(t, err)
}

func TestEncryptInvalidKey(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.pub",
		EncryptionKeyPath: "testdata/encryption.key",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.Error(t, err)

	tmpFile, err := ioutil.TempFile("/tmp/", "test-encrypt-invalid-key-")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = sopsClient.Encrypt(qualifiedDecryptedFile, tmpFile.Name())
	assert.Error(t, err)
}

func TestEncryptInvalidSourceFile(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	_, err = sopsClient.Encrypt("/invalidFile", "")
	assert.Error(t, err)
}

func TestEncryptInvalidYaml(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	_, err = sopsClient.Encrypt(invalidYamlDecryptedFile, "")
	assert.Error(t, err)
}

func TestEncryptWithRegex(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	_, err = sopsClient.Encrypt(qualifiedDecryptedFileWithRegex, "")
	assert.NoError(t, err)
}

func TestDecrypt(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	tmpFile, err := ioutil.TempFile("/tmp/", "test-decrypt-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = sopsClient.Decrypt(qualifiedEncryptedFile, tmpFile.Name())
	assert.NoError(t, err)
}

func TestDecryptInvalidKey(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.pub",
		EncryptionKeyPath: "testdata/encryption.key",
	}

	sopsClient, err := sops.NewClient(nil, options)
	require.Error(t, err)

	tmpFile, err := ioutil.TempFile("/tmp/", "test-decrypt-invalid-key-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = sopsClient.Decrypt(qualifiedEncryptedFile, tmpFile.Name())
	assert.Error(t, err)
}

func TestDecryptInvalidSrc(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	_, err = sopsClient.Decrypt("dummy", "dummy")
	assert.Error(t, err)
}

func TestDecryptMissingMetadata(t *testing.T) {
	defer deleteGpgKeys()
	options := &sops.Options{
		DecryptionKeyPath: "testdata/decryption.key",
		EncryptionKeyPath: "testdata/encryption.pub",
	}

	sopsClient, err := sops.NewClient(nil, options)
	assert.NoError(t, err)

	tmpFile, err := ioutil.TempFile("/tmp/", "test-decrypt-missing-metadata-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = sopsClient.Decrypt(missingMetadataEncryptedFile, tmpFile.Name())
	assert.Error(t, err)
}

func deleteGpgKeys() {
	gpgCmd := exec.Command("gpg", "--delete-keys", "--batch", "--yes", keyID)
	if _, err := gpgCmd.Output(); err != nil {
		fmt.Printf("error deleting key: %s\n", err)
	}

	gpgCmd = exec.Command("gpg", "--delete-secret-keys", "--batch", "--yes", keyID)
	if _, err := gpgCmd.Output(); err != nil {
		fmt.Printf("error deleting secret key: %s\n", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting home dir: %s\n", err)
		return
	}

	secRingFile := filepath.Join(homeDir, ".gnupg", "secring.gpg")
	gpgCmd = exec.Command("rm", secRingFile)
	if _, err := gpgCmd.Output(); err != nil {
		fmt.Printf("error deleting secring: %s\n", err)
	}

	return
}

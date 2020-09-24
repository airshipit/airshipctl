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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptionConfigOutputString(t *testing.T) {
	expectedEncryptionConfigYaml := `decryptionKeyPath: /tmp/decryption.pub
encryptionKeyPath: /tmp/encryption.key
`
	encryptionConfig := &EncryptionConfig{
		EncryptionKeyFileSource: EncryptionKeyFileSource{
			EncryptionKeyPath: "/tmp/encryption.key",
			DecryptionKeyPath: "/tmp/decryption.pub",
		},
	}

	assert.Equal(t, expectedEncryptionConfigYaml, encryptionConfig.String())
}

func TestValidateEncryptionConfigInvalid(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigInvalidOnlyEncKey(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		EncryptionKeyPath: "/tmp/encryption.key",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigInvalidOnlyDecKey(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		DecryptionKeyPath: "/tmp/decryption.pub",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigInvalidOnlySecretName(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		KeySecretName: "secretName",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigInvalidOnlySecretNamespace(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		KeySecretNamespace: "secretNamespace",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigValidWithSecret(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		KeySecretName:      "secretName",
		KeySecretNamespace: "secretNamespace",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

func TestValidateEncryptionConfigValidWithFile(t *testing.T) {
	encryptionConfig := &EncryptionConfigOptions{
		EncryptionKeyPath: "/tmp/encryption.key",
		DecryptionKeyPath: "/tmp/decryption.pub",
	}
	err := encryptionConfig.Validate()
	assert.Error(t, err)
}

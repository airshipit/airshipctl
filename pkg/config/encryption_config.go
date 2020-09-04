/*
Copyright 2014 The Kubernetes Authors.

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

import "sigs.k8s.io/yaml"

// EncryptionConfig holds the public and private key information
// used to encrypt and decrypt secrets
type EncryptionConfig struct {
	EncryptionKeyFileSource   `json:",inline"`
	EncryptionKeySecretSource `json:",inline"`
}

// EncryptionKeyFileSource hold the local file information for the public and private
// keys used for encryption and decryption
type EncryptionKeyFileSource struct {
	EncryptionKeyPath string `json:"encryptionKeyPath,omitempty"`
	DecryptionKeyPath string `json:"decryptionKeyPath,omitempty"`
}

// EncryptionKeySecretSource holds the secret information for the public and private
// keys used for encryption and decryption
type EncryptionKeySecretSource struct {
	KeySecretName      string `json:"keySecretName,omitempty"`
	KeySecretNamespace string `json:"keySecretNamespace,omitempty"`
}

// String returns the encryption config in yaml format
func (ec *EncryptionConfig) String() string {
	yamlData, err := yaml.Marshal(&ec)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

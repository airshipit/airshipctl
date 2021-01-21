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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/testutil"
)

const (
	encryptionConfigName  = "encryptionConfig"
	secretName            = "secretName"
	secretNamespace       = "secretNamespace"
	encryptionKeyFilePath = "/tmp/encryption.key"
	decryptionKeyFilePath = "/tmp/decryption.pub"
)

func TestConfigSetEncryptionConfigurationCmd(t *testing.T) {
	cmdTests := []*testutil.CmdTest{
		{
			Name:    "config-cmd-set-encryption-config-with-help",
			CmdLine: "--help",
			Cmd:     NewSetEncryptionConfigCommand(nil),
		},
		{
			Name:    "config-cmd-set-encryption-config-no-args",
			CmdLine: "",
			Cmd:     NewSetEncryptionConfigCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 0),
		},
		{
			Name:    "config-cmd-set-encryption-config-excess-args",
			CmdLine: "arg1 arg2",
			Cmd:     NewSetEncryptionConfigCommand(nil),
			Error:   fmt.Errorf("accepts %d arg(s), received %d", 1, 2),
		},
	}

	for _, tt := range cmdTests {
		testutil.RunTest(t, tt)
	}
}

func TestSetEncryptionConfig(t *testing.T) {
	given, cleanupGiven := testutil.InitConfig(t)
	defer cleanupGiven(t)

	tests := []struct {
		testName              string
		encryptionConfigName  string
		flags                 []string
		inputConfig           *config.Config
		secretName            string
		secretNamespace       string
		encryptionKeyFilePath string
		decryptionKeyFilePath string
		error                 error
	}{
		{
			testName:              "set-encryption-config-error-no-encryption",
			encryptionKeyFilePath: encryptionKeyFilePath,
			decryptionKeyFilePath: decryptionKeyFilePath,
			encryptionConfigName:  encryptionConfigName,
			flags: []string{
				"--decryption-key-path " + decryptionKeyFilePath,
			},
			error: fmt.Errorf("specify both encryption " +
				"and decryption keys when setting encryption config"),
			inputConfig: given,
		},
		{
			testName: "set-encryption-config-error-no-decryption",
			flags: []string{
				"--encryption-key-path " + encryptionKeyFilePath,
			},
			error: fmt.Errorf("you must specify both encryption " +
				"and decryption keys when setting encryption config"),
			encryptionConfigName:  encryptionConfigName,
			encryptionKeyFilePath: encryptionKeyFilePath,
			decryptionKeyFilePath: decryptionKeyFilePath,
		},
		{
			testName:             "set-encryption-config-error-no-options",
			encryptionConfigName: encryptionConfigName,
			error: fmt.Errorf("you must specify both encryption " +
				"and decryption keys when setting encryption config"),
			inputConfig: given,
		},
		{
			testName:              "set-encryption-config",
			encryptionConfigName:  encryptionConfigName,
			encryptionKeyFilePath: encryptionKeyFilePath,
			decryptionKeyFilePath: decryptionKeyFilePath,
			flags: []string{
				"--decryption-key-path " + decryptionKeyFilePath,
				"--encryption-key-path " + encryptionKeyFilePath,
			},
			inputConfig: given,
		},
		{
			testName:             "set-encryption-config-error-no-namespace",
			encryptionConfigName: encryptionConfigName,
			flags: []string{
				"--secret-name " + secretName,
			},
			error: fmt.Errorf("you must specify both secret name and namespace" +
				" when setting encryption config"),
		},
		{
			testName:             "set-encryption-config-error-no-secret-name",
			encryptionConfigName: encryptionConfigName,
			flags: []string{
				"--secret-namespace " + secretNamespace,
			},
			error: fmt.Errorf("you must specify both secret name and namespace" +
				" when setting encryption config"),
		},
		{
			testName:              "set-encryption-config",
			encryptionConfigName:  encryptionConfigName,
			secretName:            secretName,
			secretNamespace:       secretNamespace,
			encryptionKeyFilePath: encryptionKeyFilePath,
			decryptionKeyFilePath: decryptionKeyFilePath,
			flags: []string{
				"--secret-name " + secretName,
				"--secret-namespace " + secretNamespace,
			},
			inputConfig: given,
		},
	}

	for _, tt := range tests {
		settings := func() (*config.Config, error) {
			return tt.inputConfig, nil
		}

		cmd := &testutil.CmdTest{
			Name:    tt.testName,
			CmdLine: fmt.Sprintf("%s %s", tt.encryptionConfigName, strings.Join(tt.flags, " ")),
			Error:   tt.error,
			Cmd:     NewSetEncryptionConfigCommand(settings),
		}

		testutil.RunTest(t, cmd)

		if cmd.Error != nil {
			return
		}

		afterRunConf := tt.inputConfig
		// Find the Encryption Config Created or Modified
		afterRunEncryptionConfig, _ := afterRunConf.EncryptionConfigs[tt.encryptionConfigName]
		require.NotNil(t, afterRunEncryptionConfig)
		assert.EqualValues(t, afterRunEncryptionConfig.KeySecretName, tt.secretName)
		assert.EqualValues(t, afterRunEncryptionConfig.KeySecretNamespace, tt.secretNamespace)
		assert.EqualValues(t, afterRunEncryptionConfig.EncryptionKeyPath, tt.encryptionKeyFilePath)
		assert.EqualValues(t, afterRunEncryptionConfig.DecryptionKeyPath, tt.decryptionKeyFilePath)
	}
}

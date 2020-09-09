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
	"os"
	"path/filepath"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	kcfg "opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/secret/sops"
)

// Decrypt decrypts all encrypted files in the srcPath and writes plain text files into dstPath
func Decrypt(airconfig *config.Config, kubeconfig string,
	srcPath string, dstPath string) error {
	var cleanup kcfg.Cleanup
	var err error
	if kubeconfig == "" {
		kubeConfig := kcfg.NewBuilder().Build()
		kubeconfig, cleanup, err = kubeConfig.GetFile()
		if err != nil {
			// ignore error here and return an error later if encryption
			// config refers to secrets in apiserver
		}
		defer cleanup()
	}

	kclient, err := client.DefaultClient(airconfig.LoadedConfigPath(), kubeconfig)
	if err != nil {
		// ignore error here and return an error later if encryption
		// config refers to secrets in apiserver
	}

	return decrypt(airconfig, kclient, srcPath, dstPath)
}

func decrypt(airconfig *config.Config, kclient client.Interface, srcPath string, dstPath string) error {
	encryptionConfig, encryptionConfigErr := airconfig.CurrentContextEncryptionConfig()
	if encryptionConfigErr != nil {
		return encryptionConfigErr
	}

	if srcPath == "" {
		helper, err := phase.NewHelper(airconfig)
		if err != nil {
			return err
		}
		srcPath = helper.PhaseRoot()
	}
	if dstPath == "" {
		dstPath = srcPath
	}

	options := &sops.Options{
		KeySecretName:      encryptionConfig.KeySecretName,
		KeySecretNamespace: encryptionConfig.KeySecretNamespace,
		EncryptionKeyPath:  encryptionConfig.EncryptionKeyPath,
		DecryptionKeyPath:  encryptionConfig.DecryptionKeyPath,
	}

	sopsClient, err := sops.NewClient(kclient, options)
	if err != nil {
		return err
	}

	fs := document.NewDocumentFs()
	if fs.IsDir(srcPath) {
		absPath, _, pathErr := fs.CleanedAbs(srcPath)
		if pathErr != nil {
			return pathErr
		}

		// iterate through all files recursively and check if the directory
		// contains any secret objects with encrypt annotation
		err = fs.Walk(absPath.String(), func(encryptedFilePath string, info os.FileInfo, err error) error {
			decryptedFilePath := filepath.Join(dstPath, filepath.Base(info.Name()))
			if info.IsDir() {
				return nil
			}
			if _, err = sopsClient.Decrypt(encryptedFilePath, decryptedFilePath); err != nil {
				return err
			}
			return nil
		})
		return err
	}
	_, err = sopsClient.Decrypt(srcPath, dstPath)
	return err
}

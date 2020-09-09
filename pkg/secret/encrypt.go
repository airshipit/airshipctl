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
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	kcfg "opendev.org/airship/airshipctl/pkg/k8s/kubeconfig"
	"opendev.org/airship/airshipctl/pkg/phase"
	"opendev.org/airship/airshipctl/pkg/secret/sops"
)

const (
	encryptionFilterAnnotationKey = "airshipit.org/encrypt"
)

// Encrypt encrypts all plaintext files in the srcPath and writes encrypted files into dstPath
func Encrypt(airconfig *config.Config, kubeconfig string,
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

	return encrypt(airconfig, kclient, srcPath, dstPath)
}

func encrypt(airconfig *config.Config, kclient client.Interface, srcPath string, dstPath string) error {
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

	// if from file is directory
	fs := document.NewDocumentFs()
	if fs.IsDir(srcPath) {
		absPath, _, pathErr := fs.CleanedAbs(srcPath)
		if pathErr != nil {
			return pathErr
		}

		// iterate through all files recursively and check if the directory
		// contains any secret objects with encrypt annotation
		err = fs.Walk(absPath.String(), func(plainTextFilePath string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			// check if the file is a secret yaml
			if isValidSecret(plainTextFilePath) {
				encryptedFilePath := filepath.Join(dstPath, filepath.Base(info.Name()))
				// check if the secret has an annotation to encrypt, if not skip
				// when a directory option is passed, in place takes effect
				if _, err = sopsClient.Encrypt(plainTextFilePath, encryptedFilePath); err != nil {
					return err
				}
			}
			return nil
		})
		return err
	}

	if !isValidSecret(srcPath) {
		return nil
	}

	_, err = sopsClient.Encrypt(srcPath, dstPath)
	return err
}

// checks if the file passed is a secret object that has to be encrypted or decrypted
func isValidSecret(fileName string) bool {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return false
	}
	jsonContents, err := yaml.ToJSON(contents)
	if err != nil {
		return false
	}

	object, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsonContents)
	if err != nil {
		return false
	}
	accessor, err := meta.Accessor(object)
	if err != nil {
		return false
	}

	if accessor.GetAnnotations() != nil {
		if value, ok := accessor.GetAnnotations()[encryptionFilterAnnotationKey]; ok && value == "true" {
			return true
		}
	}

	return false
}

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

package sops

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/keys"
	"go.mozilla.org/sops/v3/keyservice"
	"go.mozilla.org/sops/v3/pgp"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"

	"opendev.org/airship/airshipctl/pkg/k8s/client"
)

const (
	tempEncryptionKeyFile         = "/tmp/encryption-key.pri"
	encryptionRegexAnnotationKey  = "airshipit.org/encryption-regex"
	encryptionFilterAnnotationKey = "airshipit.org/encrypt"
)

// Options holds the key information used to encrypt and decrypt secrets using Sops
type Options struct {
	KeySecretName      string
	KeySecretNamespace string
	EncryptionKeyPath  string
	DecryptionKeyPath  string
}

// Client is an interface that is used to encrypt and decrypt secrets
type Client interface {
	// Encrypt reads plain text secrets from srcPath and writes encrypted secrets to dstPath
	Encrypt(srcPath string, dstPath string) ([]byte, error)
	// Decrypt reads encrypted secrets from srcPath and writes plain text secrets to dstPath
	Decrypt(srcPath string, dstPath string) ([]byte, error)
}

// localGpg implements Client with local gpg keys for encryption and decryption
type localGpg struct {
	*Options
	kclient    client.Interface
	publicKey  []byte
	privateKey []byte
}

// NewClient returns a localGpg Client implementation
func NewClient(kclient client.Interface, options *Options) (Client, error) {
	client := &localGpg{
		kclient: kclient,
		Options: options,
	}
	err := client.initializeKeys()
	return client, err
}

func (lg *localGpg) initializeKeys() error {
	var publicKey, privateKey []byte
	var err error
	if lg.DecryptionKeyPath == "" && lg.EncryptionKeyPath == "" {
		// retrieve sops keys from the apiserver
		if lg.kclient == nil {
			return fmt.Errorf("kube client not initialized")
		}
		secret, apiErr := lg.getSecretFromAPI(lg.KeySecretName, lg.KeySecretNamespace)
		if apiErr != nil {
			return err
		}
		privateKey = secret.Data["pri_key"]
		publicKey = secret.Data["pub_key"]
	} else {
		// load the keys from disk
		if lg.DecryptionKeyPath != "" {
			privateKey, err = ioutil.ReadFile(lg.DecryptionKeyPath)
			if err != nil {
				return err
			}
		}
		if lg.EncryptionKeyPath != "" {
			publicKey, err = ioutil.ReadFile(lg.EncryptionKeyPath)
			if err != nil {
				return err
			}
		}
	}
	lg.publicKey = publicKey
	lg.privateKey = privateKey

	if len(lg.privateKey) > 0 {
		// import the key locally
		if err := lg.importGpgKeyPairLocally(); err != nil {
			return err
		}
	}
	return nil
}

func (lg *localGpg) importGpgKeyPairLocally() error {
	tmpPriKeyFileName := fmt.Sprintf(tempEncryptionKeyFile)

	if err := writeFile(tmpPriKeyFileName, lg.privateKey); err != nil {
		return err
	}
	defer func() {
		os.Remove(tmpPriKeyFileName)
	}()

	gpgCmd := exec.Command("gpg", "--import", tmpPriKeyFileName)
	err := gpgCmd.Run()
	if err != nil {
		return err
	}

	// gpg --export-secret-keys >~/.gnupg/secring.gpg
	// make this work with gpg1 as well for linux
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	gpgSecretImportCmd := exec.Command("gpg", "--export-secret-keys")
	secringBytes, err := gpgSecretImportCmd.Output()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(homeDir, ".gnupg", "secring.gpg"), secringBytes, 0600)
	return err
}

func (lg *localGpg) Encrypt(fromFile string, toFile string) ([]byte, error) {
	groups, err := lg.getKeyGroup(lg.publicKey)
	if err != nil {
		return nil, err
	}
	store := common.DefaultStoreForPath(fromFile)
	fileBytes, err := ioutil.ReadFile(fromFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}

	branches, err := store.LoadPlainFile(fileBytes)
	if err != nil {
		return nil, err
	}

	if err = lg.ensureNoMetadata(branches[0]); err != nil {
		// do not return error to keep this function idempotent
		// ensureNoMetadata will return an error if the file is already encrypted
		return nil, nil
	}

	// get encryption regex
	encryptionRegex, err := getEncryptionRegex(fileBytes)
	if err != nil || encryptionRegex == "" {
		encryptionRegex = "^data"
	} else if encryptionRegex != "" {
		encryptionRegex = "^data|" + encryptionRegex
	}

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			KeyGroups:      groups,
			Version:        "3.6.0",
			EncryptedRegex: encryptionRegex,
		},
		FilePath: fromFile,
	}

	keySvc := keyservice.NewLocalClient()
	dataKey, errors := tree.GenerateDataKeyWithKeyServices([]keyservice.KeyServiceClient{keySvc})
	if len(errors) > 0 {
		return nil, fmt.Errorf("%s", errors)
	}
	if err = common.EncryptTree(common.EncryptTreeOpts{
		Tree:    &tree,
		Cipher:  aes.NewCipher(),
		DataKey: dataKey,
	}); err != nil {
		return nil, err
	}

	dstStore := common.DefaultStoreForPath(toFile)
	output, err := dstStore.EmitEncryptedFile(tree)
	if err != nil {
		return nil, err
	}

	if toFile != "" {
		err = ioutil.WriteFile(toFile, output, 0600)
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func (lg *localGpg) Decrypt(fromFile string, toFile string) ([]byte, error) {
	keySvc := keyservice.NewLocalClient()
	tree, err := common.LoadEncryptedFileWithBugFixes(common.GenericDecryptOpts{
		Cipher:      aes.NewCipher(),
		InputStore:  common.DefaultStoreForPath(fromFile),
		InputPath:   fromFile,
		KeyServices: []keyservice.KeyServiceClient{keySvc},
	})
	if err != nil && err.Error() == sops.MetadataNotFound.Error() {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if _, err = common.DecryptTree(common.DecryptTreeOpts{
		Tree:        tree,
		KeyServices: []keyservice.KeyServiceClient{keySvc},
		Cipher:      aes.NewCipher(),
	}); err != nil {
		return nil, err
	}

	dstStore := common.DefaultStoreForPath(toFile)
	output, err := dstStore.EmitPlainFile(tree.Branches)
	if err != nil {
		return nil, err
	}

	if toFile != "" {
		if err = writeFile(toFile, output); err != nil {
			return nil, err
		}
	}

	return output, nil
}

// Config for generating keys.
type Config struct {
	packet.Config
	// Expiry is the duration that the generated key will be valid for.
	Expiry time.Duration
}

// Key represents an OpenPGP key.
type Key struct {
	openpgp.Entity
}

func (lg *localGpg) getSecretFromAPI(name string, namespace string) (*corev1.Secret, error) {
	return lg.kclient.ClientSet().CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
}

func (lg *localGpg) getKeyGroup(publicKeyBytes []byte) ([]sops.KeyGroup, error) {
	b := bytes.NewReader(publicKeyBytes)
	bufferedReader := bufio.NewReader(b)
	entities, err := openpgp.ReadArmoredKeyRing(bufferedReader)
	if err != nil {
		return nil, err
	}
	fingerprint := fmt.Sprintf("%X", entities[0].PrimaryKey.Fingerprint[:])
	pgpKeys := make([]keys.MasterKey, 1)
	for index, k := range pgp.MasterKeysFromFingerprintString(fingerprint) {
		pgpKeys[index] = k
	}

	var group sops.KeyGroup
	group = append(group, pgpKeys...)
	return []sops.KeyGroup{group}, nil
}

func (lg *localGpg) ensureNoMetadata(branch sops.TreeBranch) error {
	for _, b := range branch {
		if b.Key == "sops" {
			return fmt.Errorf("file already encrypted")
		}
	}
	return nil
}

func writeFile(path string, content []byte) error {
	return ioutil.WriteFile(path, content, 0600)
}

func getEncryptionRegex(yamlContent []byte) (string, error) {
	jsonContents, err := yaml.ToJSON(yamlContent)
	if err != nil {
		return "", err
	}
	object, err := runtime.Decode(unstructured.UnstructuredJSONScheme, jsonContents)
	if err != nil {
		return "", err
	}
	accessor, err := meta.Accessor(object)
	if err != nil {
		return "", err
	}

	if accessor.GetAnnotations() != nil &&
		accessor.GetAnnotations()[encryptionFilterAnnotationKey] == "true" &&
		accessor.GetAnnotations()[encryptionRegexAnnotationKey] != "" {
		return accessor.GetAnnotations()[encryptionRegexAnnotationKey], nil
	}
	return "", nil
}

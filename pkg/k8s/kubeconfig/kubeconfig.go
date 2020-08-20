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

package kubeconfig

import (
	"io"
	"log"

	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/document"
)

// Interface provides a uniform way to interact with kubeconfig file
type Interface interface {
	// GetFile returns path to kubeconfig file and a function to remove it
	// if error is returned cleanup is not needed
	GetFile() (string, Cleanup, error)
	// Write will write kubeconfig to the provided writer
	Write(w io.Writer) error
	// WriteFile will write kubeconfig data to specified path
	WriteFile(path string) error
	// WriteTempFile writes a file a temporary file, returns path to it, cleanup function and error
	// it is responsibility of the caller to use the cleanup function to make sure that there are no leftovers
	WriteTempFile(dumpRoot string) (string, Cleanup, error)
}

var _ Interface = &kubeConfig{}

type kubeConfig struct {
	path     string
	dumpRoot string

	fileSystem document.FileSystem
	sourceFunc KubeSourceFunc
}

// NewKubeConfig serves as a constructor for kubeconfig Interface
// first argument is a function that should return bytes with kubeconfig and error
// see FromByte() FromAPIalphaV1() FromFile() functions or extend with your own
// second argument are options that can be used to inject various supported options into it
// see InjectTempRoot(), InjectFileSystem(), InjectFilePath() functions for more info
func NewKubeConfig(source KubeSourceFunc, options ...Option) Interface {
	kf := &kubeConfig{}
	for _, o := range options {
		o(kf)
	}
	kf.sourceFunc = source
	if kf.fileSystem == nil {
		kf.fileSystem = document.NewDocumentFs()
	}
	return kf
}

// Option is a function that allows to modify kubeConfig object
type Option func(*kubeConfig)

// KubeSourceFunc is a function which returns bytes array to construct new kubeConfig object
type KubeSourceFunc func() ([]byte, error)

// Cleanup is a function which cleans up kubeconfig file from filesystem
type Cleanup func()

// FromByte returns KubeSource type, uses plain bytes array as source to construct kubeconfig object
func FromByte(b []byte) KubeSourceFunc {
	return func() ([]byte, error) {
		return b, nil
	}
}

// FromAPIalphaV1 returns KubeSource type, uses API Config array as source to construct kubeconfig object
func FromAPIalphaV1(apiObj *v1alpha1.KubeConfig) KubeSourceFunc {
	return func() ([]byte, error) {
		return yaml.Marshal(apiObj.Config)
	}
}

// FromFile returns KubeSource type, uses path to kubeconfig on FS as source to construct kubeconfig object
func FromFile(path string, fs document.FileSystem) KubeSourceFunc {
	return func() ([]byte, error) {
		return fs.ReadFile(path)
	}
}

// FromBundle returns KubeSource type, uses path to document bundle to find kubeconfig
func FromBundle(root string) KubeSourceFunc {
	return func() ([]byte, error) {
		docBundle, err := document.NewBundleByPath(root)
		if err != nil {
			return nil, err
		}

		config := &v1alpha1.KubeConfig{}
		selector, err := document.NewSelector().ByObject(config, v1alpha1.Scheme)
		if err != nil {
			return nil, err
		}

		doc, err := docBundle.SelectOne(selector)
		if err != nil {
			return nil, err
		}

		if err := doc.ToAPIObject(config, v1alpha1.Scheme); err != nil {
			return nil, err
		}

		return yaml.Marshal(config.Config)
	}
}

// InjectFileSystem sets fileSystem to be used, mostly to be used for tests
func InjectFileSystem(fs document.FileSystem) Option {
	return func(k *kubeConfig) {
		k.fileSystem = fs
	}
}

// InjectTempRoot sets root for temporary file system, if not set default OS temp dir will be used
func InjectTempRoot(dumpRoot string) Option {
	return func(k *kubeConfig) {
		k.dumpRoot = dumpRoot
	}
}

// InjectFilePath enables setting kubeconfig path, useful when you have kubeconfig
// from the actual filesystem, if this option is used, please also make sure that
// FromFile option is also used as a first argument in NewKubeConfig function
func InjectFilePath(path string, fs document.FileSystem) Option {
	return func(k *kubeConfig) {
		k.path = path
		k.fileSystem = fs
	}
}

func (k *kubeConfig) WriteFile(path string) (err error) {
	data, err := k.sourceFunc()
	if err != nil {
		return err
	}
	return k.fileSystem.WriteFile(path, data)
}

func (k *kubeConfig) Write(w io.Writer) (err error) {
	data, err := k.sourceFunc()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// WriteTempFile implements kubeconfig Interface
func (k *kubeConfig) WriteTempFile(root string) (string, Cleanup, error) {
	data, err := k.sourceFunc()
	if err != nil {
		return "", nil, err
	}
	file, err := k.fileSystem.TempFile(root, "kubeconfig-")
	if err != nil {
		return "", nil, err
	}
	defer file.Close()
	fName := file.Name()
	_, err = file.Write(data)
	if err != nil {
		// delete the temp file that was created and return write error
		cleanup(fName, k.fileSystem)()
		return "", nil, err
	}
	return fName, cleanup(fName, k.fileSystem), nil
}

// GetFile checks if path to kubeconfig is already set and returns it no cleanup is necessary,
// and Cleanup() method will do nothing.
// If path is not set kubeconfig will be written to temporary file system, returned path will
// point to it and Cleanup() function will remove this file from the filesystem.
func (k *kubeConfig) GetFile() (string, Cleanup, error) {
	if k.path != "" {
		return k.path, func() {}, nil
	}
	return k.WriteTempFile(k.dumpRoot)
}

func cleanup(path string, fs document.FileSystem) Cleanup {
	if path == "" {
		return func() {}
	}
	return func() {
		if err := fs.RemoveAll(path); err != nil {
			log.Fatalf("Failed to cleanup kubeconfig file %s, error: %v", path, err)
		}
	}
}

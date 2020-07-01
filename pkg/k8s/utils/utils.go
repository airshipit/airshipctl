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

package utils

import (
	"bytes"
	"io"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/cli-utils/pkg/manifestreader"

	"opendev.org/airship/airshipctl/pkg/document"
)

// FactoryFromKubeConfigPath returns a factory with the
// default Kubernetes resources for the given kube config path
func FactoryFromKubeConfigPath(kp string) cmdutil.Factory {
	kf := genericclioptions.NewConfigFlags(false)
	kf.KubeConfig = &kp
	return cmdutil.NewFactory(kf)
}

// Streams returns default IO streams object, like stdout, stdin, stderr
func Streams() genericclioptions.IOStreams {
	return genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// ManifestReaderFactory factory function for manifestreader.ManifestReader
type ManifestReaderFactory func(
	validate bool,
	bundle document.Bundle,
	factory cmdutil.Factory) manifestreader.ManifestReader

// DefaultManifestReaderFactory default factory function for manifestreader.ManifestReader
var DefaultManifestReaderFactory ManifestReaderFactory = func(
	validate bool,
	bundle document.Bundle,
	factory cmdutil.Factory) manifestreader.ManifestReader {
	return NewManifestBundleReader(validate, bundle, factory)
}

// NewManifestBundleReader returns impleemntation of manifestreader interface
func NewManifestBundleReader(
	validate bool,
	bundle document.Bundle,
	factory cmdutil.Factory) *ManifestBundleReader {
	opts := manifestreader.ReaderOptions{
		Validate:  validate,
		Namespace: metav1.NamespaceDefault,
		Factory:   factory,
	}
	buffer := bytes.NewBuffer([]byte{})
	return &ManifestBundleReader{
		Bundle: bundle,
		writer: buffer,
		StreamReader: &manifestreader.StreamManifestReader{
			ReaderName:    "airship",
			Reader:        buffer,
			ReaderOptions: opts,
		},
	}
}

// ManifestBundleReader implements manifestreader interface that to transform bundle to slice
// of *resource.Info objects using Read() method.
type ManifestBundleReader struct {
	Bundle       document.Bundle
	StreamReader *manifestreader.StreamManifestReader
	writer       io.Writer
}

func (mbr *ManifestBundleReader) Read() ([]*resource.Info, error) {
	err := mbr.Bundle.Write(mbr.writer)
	if err != nil {
		return []*resource.Info{}, err
	}
	return mbr.StreamReader.Read()
}

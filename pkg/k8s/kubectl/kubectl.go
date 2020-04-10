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

package kubectl

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

// Kubectl container holds Factory, Streams and FileSystem to
// interact with upstream kubectl objects and serves as abstraction to kubectl project
type Kubectl struct {
	cmdutil.Factory
	genericclioptions.IOStreams
	document.FileSystem
	// Directory to buffer documents before passing them to kubectl commands
	// default is empty, this means that /tmp dir will be used
	bufferDir string
}

// NewKubectl builds an instance
// of Kubectl struct from Path to kubeconfig file
func NewKubectl(f cmdutil.Factory) *Kubectl {
	return &Kubectl{
		Factory: f,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		FileSystem: document.NewDocumentFs(),
	}
}

func (kubectl *Kubectl) WithBufferDir(bd string) *Kubectl {
	kubectl.bufferDir = bd
	return kubectl
}

// Apply is abstraction to kubectl apply command
func (kubectl *Kubectl) Apply(docs []document.Document, ao *ApplyOptions) error {
	tf, err := kubectl.TempFile(kubectl.bufferDir, "initinfra")
	if err != nil {
		return err
	}

	defer func(f document.File) {
		fName := f.Name()
		dErr := kubectl.RemoveAll(fName)
		if dErr != nil {
			log.Fatalf("Failed to cleanup temporary file %s during kubectl apply", fName)
		}
	}(tf)
	defer tf.Close()
	for _, doc := range docs {
		// Write out documents to temporary file
		err = utilyaml.WriteOut(tf, doc)
		if err != nil {
			return err
		}
	}
	ao.SetSourceFiles([]string{tf.Name()})
	return ao.Run()
}

// ApplyOptions is a wrapper over kubectl ApplyOptions, used to build
// new options from the factory and iostreams defined in Kubectl container
func (kubectl *Kubectl) ApplyOptions() (*ApplyOptions, error) {
	return NewApplyOptions(kubectl.Factory, kubectl.IOStreams)
}

package kubectl

import (
	"os"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/kustomize/v3/pkg/fs"

	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	utilyaml "opendev.org/airship/airshipctl/pkg/util/yaml"
)

// Kubectl container holds Factory, Streams and FileSystem to
// interact with upstream kubectl objects and serves as abstraction to kubectl project
type Kubectl struct {
	cmdutil.Factory
	genericclioptions.IOStreams
	FileSystem
	// Directory to buffer documents before passing them to kubectl commands
	// default is empty, this means that /tmp dir will be used
	bufferDir string
}

// NewKubectlFromKubeconfigPath builds an instance
// of Kubectl struct from Path to kubeconfig file
func NewKubectl(f cmdutil.Factory) *Kubectl {
	return &Kubectl{
		Factory: f,
		IOStreams: genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
		FileSystem: Buffer{FileSystem: fs.MakeRealFS()},
	}
}

func (kubectl *Kubectl) WithBufferDir(bd string) *Kubectl {
	kubectl.bufferDir = bd
	return kubectl
}

// Apply is abstraction to kubectl apply command
func (kubectl *Kubectl) Apply(docs []document.Document, ao *apply.ApplyOptions) error {
	tf, err := kubectl.TempFile(kubectl.bufferDir, "initinfra")
	if err != nil {
		return err
	}

	defer func(f File) {
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
	ao.DeleteOptions.Filenames = []string{tf.Name()}
	return ao.Run()
}

// ApplyOptions is a wrapper over kubectl ApplyOptions, used to build
// new options from the factory and iostreams defined in Kubectl container
func (kubectl *Kubectl) ApplyOptions() (*apply.ApplyOptions, error) {
	return NewApplyOptions(kubectl.Factory, kubectl.IOStreams)
}

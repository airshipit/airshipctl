package kubectl

import (
	"io/ioutil"

	"sigs.k8s.io/kustomize/v3/pkg/fs"
)

// File extends kustomize File and provide abstraction to creating temporary files
type File interface {
	fs.File
	Name() string
}

// FileSystem extends kustomize FileSystem and provide abstraction to creating temporary files
type FileSystem interface {
	fs.FileSystem
	TempFile(string, string) (File, error)
}

// Buffer is adaptor to TempFile
type Buffer struct {
	fs.FileSystem
}

// TempFile creates file in temporary filesystem, at default os.TempDir
func (b Buffer) TempFile(tmpDir string, prefix string) (File, error) {
	return ioutil.TempFile(tmpDir, prefix)
}

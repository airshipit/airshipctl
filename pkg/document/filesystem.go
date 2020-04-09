package document

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

// DocumentFs is adaptor to TempFile
type DocumentFs struct {
	fs.FileSystem
}

// NewDocumentFs returns an instalce of DocumentFs
func NewDocumentFs() FileSystem {
	return &DocumentFs{FileSystem: fs.MakeFsOnDisk()}
}

// TempFile creates file in temporary filesystem, at default os.TempDir
func (dfs DocumentFs) TempFile(tmpDir string, prefix string) (File, error) {
	return ioutil.TempFile(tmpDir, prefix)
}

package repo

import (
	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage"
)

// Adapter is abstraction to SVC
type Adapter interface {
	Open() error
	Clone(co *git.CloneOptions) error
	Fetch(fo *git.FetchOptions) error
	Worktree() (*git.Worktree, error)
	Head() (*plumbing.Reference, error)
	ResolveRevision(plumbing.Revision) (*plumbing.Hash, error)
	IsOpen() bool
	SetFilesystem(billy.Filesystem)
	SetStorer(s storage.Storer)
	Close()
}

// GitDriver implements repository interface
type GitDriver struct {
	*git.Repository
	Filesystem billy.Filesystem
	Storer     storage.Storer
}

func NewGitDriver(fs billy.Filesystem, s storage.Storer) Adapter {
	return &GitDriver{Storer: s, Filesystem: fs}
}

// Open implements repository interface
func (g *GitDriver) Open() error {
	r, err := git.Open(g.Storer, g.Filesystem)
	if err != nil {
		return err
	}
	g.Repository = r
	return nil
}

func (g *GitDriver) IsOpen() bool {
	if g.Repository == nil {
		return false
	}
	return true
}

// Close sets repository to nil, IsOpen() function will return false now
func (g *GitDriver) Close() {
	g.Repository = nil
}

// Clone implements repository interface
func (g *GitDriver) Clone(co *git.CloneOptions) error {
	r, err := git.Clone(g.Storer, g.Filesystem, co)
	if err != nil {
		return err
	}
	g.Repository = r
	return nil
}

func (g *GitDriver) SetFilesystem(fs billy.Filesystem) {
	g.Filesystem = fs
}

func (g *GitDriver) SetStorer(s storage.Storer) {
	g.Storer = s
}

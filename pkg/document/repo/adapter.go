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

package repo

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage"
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

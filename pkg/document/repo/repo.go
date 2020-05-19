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
	"errors"
	"fmt"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"

	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

//TODO(sy495p): Switch to strongly-typed errors in this file

// variables for repository errors
var (
	ErrNoOpenRepo   = errors.New("no open repository is stored")
	ErrCantParseURL = errors.New("could not get target directory from url")
)

// OptionsBuilder interface provides specification for a repository implementation
type OptionsBuilder interface {
	ToAuth() (transport.AuthMethod, error)
	ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions
	ToCheckoutOptions(force bool) *git.CheckoutOptions
	ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions
	URL() string
}

// Repository container holds Filesystem, spec and open repository object
// Abstracts git repository and allows for easy cloning, checkout and update of git repos
type Repository struct {
	Driver Adapter
	OptionsBuilder
	Name string
}

// NewRepository create repository object, with real filesystem on disk
// basePath is used to calculate final path where to clone/open the repository
func NewRepository(basePath string, builder OptionsBuilder) (*Repository, error) {
	dirName := util.GitDirNameFromURL(builder.URL())
	if dirName == "" {
		return nil, fmt.Errorf("URL: %s, original error: %w", builder.URL(), ErrCantParseURL)
	}
	fs := osfs.New(filepath.Join(basePath, dirName))

	s, err := storerFromFs(fs)
	if err != nil {
		return nil, err
	}

	// This can create
	return &Repository{
		Name:           dirName,
		Driver:         NewGitDriver(fs, s),
		OptionsBuilder: builder,
	}, nil
}

func storerFromFs(fs billy.Filesystem) (storage.Storer, error) {
	dot, err := fs.Chroot(".git")
	if err != nil {
		return nil, err
	}
	return filesystem.NewStorage(dot, cache.NewObjectLRUDefault()), nil
}

// Update fetches new refs, and checkout according to checkout options
func (repo *Repository) Update(force bool) error {
	log.Debugf("Updating repository %s", repo.Name)
	if !repo.Driver.IsOpen() {
		return ErrNoOpenRepo
	}
	auth, err := repo.ToAuth()
	if err != nil {
		return err
	}
	err = repo.Driver.Fetch(repo.ToFetchOptions(auth))
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch refs for repository %v: %w", repo.Name, err)
	}
	return repo.Checkout(force)
}

// Checkout git repository, ToCheckoutOptions method will be used go get CheckoutOptions
func (repo *Repository) Checkout(enforce bool) error {
	log.Debugf("Attempting to checkout the repository %s", repo.Name)
	if !repo.Driver.IsOpen() {
		return ErrNoOpenRepo
	}
	co := repo.ToCheckoutOptions(enforce)
	tree, err := repo.Driver.Worktree()
	if err != nil {
		return fmt.Errorf("could not get worktree from the repo, %w", err)
	}
	return tree.Checkout(co)
}

// Open the repository
func (repo *Repository) Open() error {
	log.Debugf("Attempting to open repository %s", repo.Name)
	return repo.Driver.Open()
}

// Clone given repository
func (repo *Repository) Clone() error {
	log.Debugf("Attempting to clone the repository %s", repo.Name)
	auth, err := repo.ToAuth()
	if err != nil {
		return fmt.Errorf("failed to build auth options for repository %v: %w", repo.Name, err)
	}

	return repo.Driver.Clone(repo.ToCloneOptions(auth))
}

// Download will clone and checkout repository based on auth and checkout fields of the Repository object
// If repository is already cloned, it will be opened and checked out to configured hash,branch,tag etc...
// no remotes will be modified in this case, also no refs will be updated.
// enforce parameter is used to simulate git reset --hard option.
// If you want to enforce state of the repository, please delete current git repository before downloading.
func (repo *Repository) Download(enforceCheckout bool) error {
	log.Debugf("Attempting to download the repository %s", repo.Name)

	if !repo.Driver.IsOpen() {
		err := repo.Clone()
		if err == git.ErrRepositoryAlreadyExists {
			openErr := repo.Open()
			if openErr != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	return repo.Checkout(enforceCheckout)
}

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

package config

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/errors"
	"opendev.org/airship/airshipctl/pkg/log"
)

// Constants for possible repo authentication types
const (
	SSHAuth   = "ssh-key"
	SSHPass   = "ssh-pass"
	HTTPBasic = "http-basic"
)

// remoteName is a remote name that airshipctl work with during document pull
// TODO (raliev) consider make this variable configurable via repoCheckout options
var remoteName = git.DefaultRemoteName

// Repository struct holds the information for the remote sources of manifest yaml documents.
// Information such as location, authentication info,
// as well as details of what to get such as branch, tag, commit it, etc.
type Repository struct {
	// URLString for Repository
	URLString string `json:"url"`
	// Auth holds authentication options against remote
	Auth *RepoAuth `json:"auth,omitempty"`
	// CheckoutOptions holds options to checkout repository
	CheckoutOptions *RepoCheckout `json:"checkout,omitempty"`
	// FetchOptions holds options for fetching remote refs
	FetchOptions *RepoFetch `json:"fetch,omitempty"`
}

// RepoAuth struct describes method of authentication against given repository
type RepoAuth struct {
	// Type of authentication method to be used with given repository
	// supported types are "ssh-key", "ssh-pass", "http-basic"
	Type string `json:"type,omitempty"`
	//KeyPassword is a password decrypt ssh private key (used with ssh-key auth type)
	KeyPassword string `json:"keyPass,omitempty"`
	// KeyPath is path to private ssh key on disk (used with ssh-key auth type)
	KeyPath string `json:"sshKey,omitempty"`
	//HTTPPassword is password for basic http authentication (used with http-basic auth type)
	HTTPPassword string `json:"httpPass,omitempty"`
	// SSHPassword is password for ssh password authentication (used with ssh-pass)
	SSHPassword string `json:"sshPass,omitempty"`
	// Username to authenticate against git remote (used with any type)
	Username string `json:"username,omitempty"`
}

// RepoCheckout container holds information how to checkout repository
// Each field is mutually exclusive
type RepoCheckout struct {
	// CommitHash is full hash of the commit that will be used to checkout
	CommitHash string `json:"commitHash"`
	// Branch is the branch name to checkout
	Branch string `json:"branch"`
	// Tag is the tag name to checkout
	Tag string `json:"tag"`
	// Ref is the ref to checkout
	// for example refs/changes/04/691202/5
	Ref string `json:"ref,omitempty"`
	// ForceCheckout is a boolean to indicate whether to use the `--force` option when checking out
	ForceCheckout bool `json:"force"`
	// LocalBranch is a boolean to indicate whether the Branch is local one. False by default
	LocalBranch bool `json:"localBranch"`
}

// RepoFetch holds information on which remote ref to fetch
type RepoFetch struct {
	// RemoteRefSpec is used for remote fetches such as gerrit change
	// requests and github pull requests. The format of the refspec is an
	// optional +, followed by <src>:<dst>, where <src> is the pattern for
	// references on the remote side and <dst> is where those references
	// will be written locally. The + tells Git to update the reference
	// even if it isn't a fast-forward.
	// eg.: refs/changes/04/691202/5:refs/changes/04/691202/5
	RemoteRefSpec string `json:"remoteRefSpec,omitempty"`
}

// RepoCheckout methods

func (c *RepoCheckout) String() string {
	yaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Validate checks for possible values for
// repository checkout and returns Error for incorrect values
// returns nil when there are no errors
func (c *RepoCheckout) Validate() error {
	possibleValues := []string{c.CommitHash, c.Branch, c.Tag, c.Ref}
	var count int
	for _, val := range possibleValues {
		if val != "" {
			count++
		}
	}
	if count > 1 {
		return ErrMutuallyExclusiveCheckout{}
	}
	return nil
}

// Validate verifies that the remote refspec is valid. If a remote refspec was
// not supplied, Validate does nothing.
func (rf *RepoFetch) Validate() error {
	if rf.RemoteRefSpec != "" {
		return gitconfig.RefSpec(rf.RemoteRefSpec).Validate()
	}
	return nil
}

// RepoAuth methods
var (
	AllowedAuthTypes = []string{SSHAuth, SSHPass, HTTPBasic}
)

// String returns repository authentication details in string format
func (auth *RepoAuth) String() string {
	yaml, err := yaml.Marshal(&auth)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Validate checks for possible values for
// repository authentication and returns Error for incorrect values
// returns nil when there are no errors
func (auth *RepoAuth) Validate() error {
	if !stringInSlice(auth.Type, AllowedAuthTypes) {
		return ErrAuthTypeNotSupported{}
	}

	switch auth.Type {
	case SSHAuth:
		if auth.HTTPPassword != "" || auth.SSHPassword != "" {
			return NewErrIncompatibleAuthOptions([]string{"http-pass, ssh-pass"}, auth.Type)
		}
	case HTTPBasic:
		if auth.SSHPassword != "" || auth.KeyPath != "" || auth.KeyPassword != "" {
			return NewErrIncompatibleAuthOptions([]string{"ssh-pass, ssh-key, key-pass"}, auth.Type)
		}
	case SSHPass:
		if auth.KeyPath != "" || auth.KeyPassword != "" || auth.HTTPPassword != "" {
			return NewErrIncompatibleAuthOptions([]string{"ssh-key, key-pass, http-pass"}, auth.Type)
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// String returns repository details in a string format
func (repo *Repository) String() string {
	yaml, err := yaml.Marshal(&repo)
	if err != nil {
		return ""
	}
	return string(yaml)
}

// Validate check possible values for repository and
// returns Error when incorrect value is given
// returns nil when there are no errors
func (repo *Repository) Validate() error {
	if repo.URLString == "" {
		return ErrRepoSpecRequiresURL{}
	}

	if repo.Auth != nil {
		err := repo.Auth.Validate()
		if err != nil {
			return err
		}
	}

	if repo.CheckoutOptions != nil {
		err := repo.CheckoutOptions.Validate()
		if err != nil {
			return err
		}
	} else {
		log.Debugf("Checkout options not defined, cloning from master")
	}

	return nil
}

// ToAuth returns an implementation of transport.AuthMethod for
// the given auth type to establish an ssh connection
func (repo *Repository) ToAuth() (transport.AuthMethod, error) {
	if repo.Auth == nil {
		return nil, nil
	}
	switch repo.Auth.Type {
	case SSHAuth:
		return ssh.NewPublicKeysFromFile(repo.Auth.Username, repo.Auth.KeyPath, repo.Auth.KeyPassword)
	case SSHPass:
		return &ssh.Password{User: repo.Auth.Username, Password: repo.Auth.HTTPPassword}, nil
	case HTTPBasic:
		return &http.BasicAuth{Username: repo.Auth.Username, Password: repo.Auth.HTTPPassword}, nil
	default:
		return nil, errors.ErrNotImplemented{What: fmt.Sprintf("authtype %s", repo.Auth.Type)}
	}
}

// ToCheckoutOptions returns an instance of git.CheckoutOptions with
// respective values(Branch/Tag/Hash) in checkout options initialized
// CheckoutOptions describes how a checkout operation should be performed
func (repo *Repository) ToCheckoutOptions() *git.CheckoutOptions {
	co := &git.CheckoutOptions{}
	if repo.CheckoutOptions != nil {
		co.Force = repo.CheckoutOptions.ForceCheckout
		switch {
		case repo.CheckoutOptions.Branch != "":
			if repo.CheckoutOptions.LocalBranch {
				co.Branch = plumbing.NewBranchReferenceName(repo.CheckoutOptions.Branch)
			} else {
				co.Branch = plumbing.NewRemoteReferenceName(remoteName, repo.CheckoutOptions.Branch)
			}
		case repo.CheckoutOptions.Tag != "":
			co.Branch = plumbing.NewTagReferenceName(repo.CheckoutOptions.Tag)
		case repo.CheckoutOptions.CommitHash != "":
			co.Hash = plumbing.NewHash(repo.CheckoutOptions.CommitHash)
		case repo.CheckoutOptions.Ref != "":
			co.Branch = plumbing.ReferenceName(repo.CheckoutOptions.Ref)
		}
	}
	return co
}

// ToCloneOptions returns an instance of git.CloneOptions with
// authentication and URL set
// CloneOptions describes how a clone should be performed
func (repo *Repository) ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions {
	return &git.CloneOptions{
		Auth:       auth,
		URL:        repo.URLString,
		RemoteName: remoteName,
	}
}

// ToFetchOptions returns an instance of git.FetchOptions for given authentication
// FetchOptions describes how a fetch should be performed
func (repo *Repository) ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions {
	var refSpecs []gitconfig.RefSpec
	if repo.FetchOptions != nil && repo.FetchOptions.RemoteRefSpec != "" {
		refSpecs = []gitconfig.RefSpec{gitconfig.RefSpec(repo.FetchOptions.RemoteRefSpec)}
	}
	return &git.FetchOptions{
		Auth:     auth,
		RefSpecs: refSpecs,
	}
}

// URL returns the repository URL in a string format
func (repo *Repository) URL() string {
	return repo.URLString
}

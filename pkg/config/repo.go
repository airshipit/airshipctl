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
	possibleValues := []string{c.CommitHash, c.Branch, c.Tag, c.RemoteRef}
	var count int
	for _, val := range possibleValues {
		if val != "" {
			count++
		}
	}
	if count > 1 {
		return ErrMutuallyExclusiveCheckout{}
	}
	if c.RemoteRef != "" {
		return errors.ErrNotImplemented{What: "repository checkout by RemoteRef"}
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
			return NewErrIncompetibleAuthOptions([]string{"http-pass, ssh-pass"}, auth.Type)
		}
	case HTTPBasic:
		if auth.SSHPassword != "" || auth.KeyPath != "" || auth.KeyPassword != "" {
			return NewErrIncompetibleAuthOptions([]string{"ssh-pass, ssh-key, key-pass"}, auth.Type)
		}
	case SSHPass:
		if auth.KeyPath != "" || auth.KeyPassword != "" || auth.HTTPPassword != "" {
			return NewErrIncompetibleAuthOptions([]string{"ssh-key, key-pass, http-pass"}, auth.Type)
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
func (repo *Repository) ToCheckoutOptions(force bool) *git.CheckoutOptions {
	co := &git.CheckoutOptions{
		Force: force,
	}
	if repo.CheckoutOptions != nil {
		switch {
		case repo.CheckoutOptions.Branch != "":
			co.Branch = plumbing.NewBranchReferenceName(repo.CheckoutOptions.Branch)
		case repo.CheckoutOptions.Tag != "":
			co.Branch = plumbing.NewTagReferenceName(repo.CheckoutOptions.Tag)
		case repo.CheckoutOptions.CommitHash != "":
			co.Hash = plumbing.NewHash(repo.CheckoutOptions.CommitHash)
		}
	}
	return co
}

// ToCloneOptions returns an instance of git.CloneOptions with
// authentication and URL set
// CloneOptions describes how a clone should be performed
func (repo *Repository) ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions {
	cl := &git.CloneOptions{
		Auth: auth,
		URL:  repo.URLString,
	}
	if repo.CheckoutOptions != nil {
		switch {
		case repo.CheckoutOptions.Branch != "":
			cl.ReferenceName = plumbing.NewBranchReferenceName(repo.CheckoutOptions.Branch)
		case repo.CheckoutOptions.Tag != "":
			cl.ReferenceName = plumbing.NewTagReferenceName(repo.CheckoutOptions.Tag)
		}
	}
	return cl
}

// ToFetchOptions returns an instance of git.FetchOptions for given authentication
// FetchOptions describes how a fetch should be performed
func (repo *Repository) ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions {
	return &git.FetchOptions{Auth: auth}
}

// URL returns the repository URL in a string format
func (repo *Repository) URL() string {
	return repo.URLString
}

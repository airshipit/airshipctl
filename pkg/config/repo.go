package config

import (
	"fmt"
	"reflect"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"sigs.k8s.io/yaml"

	"opendev.org/airship/airshipctl/pkg/errors"
)

const (
	SSHAuth   = "ssh-key"
	SSHPass   = "ssh-pass"
	HTTPBasic = "http-basic"
)

// RepoCheckout methods

func (c *RepoCheckout) Equal(s *RepoCheckout) bool {
	if s == nil {
		return s == c
	}
	return c.CommitHash == s.CommitHash &&
		c.Branch == s.Branch &&
		c.Tag == s.Tag &&
		c.RemoteRef == s.RemoteRef
}

func (c *RepoCheckout) String() string {
	yaml, err := yaml.Marshal(&c)
	if err != nil {
		return ""
	}
	return string(yaml)
}

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
		return fmt.Errorf("Repository checkout by RemoteRef is not yet implemented\n%w", errors.ErrNotImplemented{})
	}
	return nil
}

// RepoAuth methods
var (
	AllowedAuthTypes = []string{SSHAuth, SSHPass, HTTPBasic}
)

func (auth *RepoAuth) Equal(s *RepoAuth) bool {
	if s == nil {
		return s == auth
	}
	return auth.Type == s.Type &&
		auth.KeyPassword == s.KeyPassword &&
		auth.KeyPath == s.KeyPath &&
		auth.SSHPassword == s.SSHPassword &&
		auth.Username == s.Username
}

func (auth *RepoAuth) String() string {
	yaml, err := yaml.Marshal(&auth)
	if err != nil {
		return ""
	}
	return string(yaml)
}

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

// Repository functions
// Equal compares repository specs
func (repo *Repository) Equal(s *Repository) bool {
	if s == nil {
		return s == repo
	}

	return repo.URLString == s.URLString &&
		reflect.DeepEqual(s.Auth, repo.Auth) &&
		reflect.DeepEqual(s.CheckoutOptions, repo.CheckoutOptions)
}

func (repo *Repository) String() string {
	yaml, err := yaml.Marshal(&repo)
	if err != nil {
		return ""
	}
	return string(yaml)
}

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
	}

	return nil
}

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
		return nil, fmt.Errorf("Error building auth opts, repo\n%s\n: %w", repo.String(), errors.ErrNotImplemented{})
	}
}

func (repo *Repository) ToCheckoutOptions(force bool) *git.CheckoutOptions {
	co := &git.CheckoutOptions{
		Force: force,
	}
	switch {
	case repo.CheckoutOptions == nil:
	case repo.CheckoutOptions.Branch != "":
		co.Branch = plumbing.NewBranchReferenceName(repo.CheckoutOptions.Branch)
	case repo.CheckoutOptions.Tag != "":
		co.Branch = plumbing.NewTagReferenceName(repo.CheckoutOptions.Tag)
	case repo.CheckoutOptions.CommitHash != "":
		co.Hash = plumbing.NewHash(repo.CheckoutOptions.CommitHash)
	}
	return co
}

func (repo *Repository) ToCloneOptions(auth transport.AuthMethod) *git.CloneOptions {
	return &git.CloneOptions{
		Auth: auth,
		URL:  repo.URLString,
	}
}

func (repo *Repository) ToFetchOptions(auth transport.AuthMethod) *git.FetchOptions {
	return &git.FetchOptions{Auth: auth}
}

func (repo *Repository) URL() string {
	return repo.URLString
}

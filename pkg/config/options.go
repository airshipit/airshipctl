/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"os"

	"opendev.org/airship/airshipctl/pkg/errors"
)

// AuthInfoOptions holds all configurable options for
// authentication information or credential
type AuthInfoOptions struct {
	Name              string
	ClientCertificate string
	ClientKey         string
	Token             string
	Username          string
	Password          string
	EmbedCertData     bool
}

// ContextOptions holds all configurable options for context
type ContextOptions struct {
	Name           string
	ClusterType    string
	CurrentContext bool
	Cluster        string
	AuthInfo       string
	Manifest       string
	Namespace      string
	Current        bool
}

// ClusterOptions holds all configurable options for cluster configuration
type ClusterOptions struct {
	Name                  string
	ClusterType           string
	Server                string
	InsecureSkipTLSVerify bool
	CertificateAuthority  string
	EmbedCAData           bool
}

// ManifestOptions holds all configurable options for manifest configuration
type ManifestOptions struct {
	Name       string
	RepoName   string
	URL        string
	Branch     string
	CommitHash string
	Tag        string
	RemoteRef  string
	Force      bool
	IsPrimary  bool
	SubPath    string
	TargetPath string
}

// TODO(howell): The following functions are tightly coupled with flags passed
// on the command line. We should find a way to remove this coupling, since it
// is possible to create (and validate) these objects without using the command
// line.

// TODO(howell): strongly type the errors in this file

// Validate checks for the possible authentication values and returns
// Error when invalid value or incompatible choice of values given
func (o *AuthInfoOptions) Validate() error {
	// TODO(howell): This prevents a user of airshipctl from creating a
	// credential with both a bearer-token and a user/password, but it does
	// not prevent a user from adding a bearer-token to a credential which
	// already had a user/pass and visa-versa. This could create bugs if a
	// user at first chooses one method, but later switches to another.
	if o.Token != "" && (o.Username != "" || o.Password != "") {
		return ErrConflictingAuthOptions{}
	}

	if !o.EmbedCertData {
		return nil
	}

	if err := checkExists("client-certificate", o.ClientCertificate); err != nil {
		return err
	}

	if err := checkExists("client-key", o.ClientKey); err != nil {
		return err
	}

	return nil
}

// Validate checks for the possible context option values and returns
// Error when invalid value or incompatible choice of values given
func (o *ContextOptions) Validate() error {
	if !o.Current && o.Name == "" {
		return ErrEmptyContextName{}
	}

	if o.Current && o.Name != "" {
		return ErrConflictingContextOptions{}
	}

	// If the user simply wants to change the current context, no further validation is needed
	if o.CurrentContext {
		return nil
	}

	// If the cluster-type was specified, verify that it's valid
	if o.ClusterType != "" {
		if err := ValidClusterType(o.ClusterType); err != nil {
			return err
		}
	}

	// TODO Manifest, Cluster could be validated against the existing config maps
	return nil
}

// Validate checks for the possible cluster option values and returns
// Error when invalid value or incompatible choice of values given
func (o *ClusterOptions) Validate() error {
	if o.Name == "" {
		return ErrEmptyClusterName{}
	}

	err := ValidClusterType(o.ClusterType)
	if err != nil {
		return err
	}

	if o.InsecureSkipTLSVerify && o.CertificateAuthority != "" {
		return ErrConflictingClusterOptions{}
	}

	if !o.EmbedCAData {
		return nil
	}

	if err := checkExists("certificate-authority", o.CertificateAuthority); err != nil {
		return err
	}

	return nil
}

func checkExists(flagName, path string) error {
	if path == "" {
		return fmt.Errorf("you must specify a --%s to embed", flagName)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("could not read %s data from '%s': %v", flagName, path, err)
	}
	return nil
}

// Validate checks for the possible manifest option values and returns
// Error when invalid value or incompatible choice of values given
func (o *ManifestOptions) Validate() error {
	if o.Name == "" {
		return fmt.Errorf("you must specify a non-empty Manifest name")
	}
	if o.RemoteRef != "" {
		return fmt.Errorf("Repository checkout by RemoteRef is not yet implemented\n%w", errors.ErrNotImplemented{})
	}
	if o.IsPrimary && o.RepoName == "" {
		return ErrMissingRepositoryName{}
	}
	possibleValues := [3]string{o.CommitHash, o.Branch, o.Tag}
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

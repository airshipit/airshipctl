/*
Copyright 2014 The Kubernetes Authors.

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

import "sigs.k8s.io/yaml"

// Manifest is a tuple of references to a Manifest (how do Identify, collect ,
// find the yaml manifests that airship uses to perform its operations)
type Manifest struct {
	// PrimaryRepositoryName is a name of the repo, that contains site/<site-name> directory
	// and is a starting point for building document bundle
	PrimaryRepositoryName string `json:"primaryRepositoryName"`
	// ExtraRepositories is the map of extra repositories addressable by a name
	Repositories map[string]*Repository `json:"repositories,omitempty"`
	// TargetPath Local Target path for working or home directory for all Manifest Cloned/Returned/Generated
	TargetPath string `json:"targetPath"`
	// SubPath is a path relative to TargetPath + Path where PrimaryRepository is cloned and contains
	// directories with ClusterType and Phase bundles, example:
	// Repositories[PrimaryRepositoryName].Url = 'https://github.com/airshipit/treasuremap'
	// SubPath = "manifests"
	// you would expect that at treasuremap/manifests you would have ephemeral/initinfra and
	// ephemera/target directories, containing kustomize.yaml.
	SubPath string `json:"subPath"`
	// MetadataPath path to a metadata file relative to TargetPath
	MetadataPath string `json:"metadataPath"`
}

// Repository is a tuple that holds the information for the remote sources of manifest yaml documents.
// Information such as location, authentication info,
// as well as details of what to get such as branch, tag, commit it, etc.
type Repository struct {
	// URLString for Repository
	URLString string `json:"url"`
	// Auth holds authentication options against remote
	Auth *RepoAuth `json:"auth,omitempty"`
	// CheckoutOptions holds options to checkout repository
	CheckoutOptions *RepoCheckout `json:"checkout,omitempty"`
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
	// RemoteRef is not supported currently TODO
	// RemoteRef is used for remote checkouts such as gerrit change requests/github pull request
	// for example refs/changes/04/691202/5
	// TODO Add support for fetching remote refs
	RemoteRef string `json:"remoteRef,omitempty"`
	// ForceCheckout is a boolean to indicate whether to use the `--force` option when checking out
	ForceCheckout bool `json:"force"`
}

// Metadata holds entrypoints for phases, inventory and clusterctl
type Metadata struct {
	Inventory *InventoryMeta `json:"inventory,omitempty"`
	PhaseMeta *PhaseMeta     `json:"phase,omitempty"`
}

// InventoryMeta holds inventory metadata, this is to be extended in the future
// when we have more information how to handle non-baremetal inventories
// path is a kustomize entrypoint against which we will build bundle containing bmh hosts
type InventoryMeta struct {
	Path string `json:"path,omitempty"`
}

// PhaseMeta holds phase metadata, right now it is only path, but maybe extended further
// path is a kustomize entrypoint against which we will build bundle with phase objects
type PhaseMeta struct {
	Path string `json:"path,omitempty"`
}

// Manifest functions
func (m *Manifest) String() string {
	yamlData, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

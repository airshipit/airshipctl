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

import (
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"
)

// Where possible, json tags match the cli argument names.
// Top level config objects and all values required for proper functioning are not "omitempty".
// Any truly optional piece of config is allowed to be omitted.

// Config holds the information required by airshipctl commands
// It is somewhat a superset of what a kubeconfig looks like, we allow for this overlaps by providing
// a mechanism to consume or produce a kubeconfig into / from the airship config.
type Config struct {
	// +optional
	Kind string `json:"kind,omitempty"`

	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Clusters is a map of referenceable names to cluster configs
	Clusters map[string]*ClusterPurpose `json:"clusters"`

	// AuthInfos is a map of referenceable names to user configs
	AuthInfos map[string]*AuthInfo `json:"users"`

	// Contexts is a map of referenceable names to context configs
	Contexts map[string]*Context `json:"contexts"`

	// Manifests is a map of referenceable names to documents
	Manifests map[string]*Manifest `json:"manifests"`

	// CurrentContext is the name of the context that you would like to use by default
	CurrentContext string `json:"current-context"`

	// Modules Section
	// Will store configuration required by the different airshipctl modules
	// Such as Bootstrap, Workflows, Document, etc
	ModulesConfig *Modules `json:"modules-config"`

	// loadedConfigPath is the full path to the the location of the config
	// file from which this config was loaded
	// +not persisted in file
	loadedConfigPath string

	// kubeConfigPath is the full path to the the location of the
	// kubeconfig file associated with this airship config instance
	// +not persisted in file
	kubeConfigPath string

	// Private instance of Kube Config content as an object
	kubeConfig *kubeconfig.Config
}

// Encapsulates the Cluster Type as an enumeration
type ClusterPurpose struct {
	// Cluster map of referenceable names to cluster configs
	ClusterTypes map[string]*Cluster `json:"cluster-type"`
}

// Cluster contains information about how to communicate with a kubernetes cluster
type Cluster struct {
	// Complex cluster name defined by the using <cluster name>_<cluster type>)
	NameInKubeconf string `json:"cluster-kubeconf"`

	// KubeConfig Cluster Object
	cluster *kubeconfig.Cluster

	// Bootstrap configuration this clusters ephemeral hosts will rely on
	Bootstrap string `json:"bootstrap-info"`
}

// Modules, generic configuration for modules
// Configuration that the Bootstrap Module would need
// Configuration that the Document Module would need
// Configuration that the Workflows Module would need
type Modules struct {
	BootstrapInfo map[string]*Bootstrap `json:"bootstrapInfo"`
}

// Context is a tuple of references to a cluster (how do I communicate with a kubernetes context),
// a user (how do I identify myself), and a namespace (what subset of resources do I want to work with)
type Context struct {
	// Context name in kubeconf
	NameInKubeconf string `json:"context-kubeconf"`

	// Manifest is the default manifest to be use with this context
	// +optional
	Manifest string `json:"manifest,omitempty"`

	// KubeConfig Context Object
	context *kubeconfig.Context
}

type AuthInfo struct {
	// KubeConfig AuthInfo Object
	authInfo *kubeconfig.AuthInfo
}

// Manifest is a tuple of references to a Manifest (how do Identify, collect ,
// find the yaml manifests that airship uses to perform its operations)
type Manifest struct {
	// PrimaryRepositoryName is a name of the repo, that contains site/<site-name> directory
	// and is a starting point for building document bundle
	PrimaryRepositoryName string `json:"primary-repository-name"`
	// ExtraRepositories is the map of extra repositories addressable by a name
	Repositories map[string]*Repository `json:"repositories,omitempty"`
	// TargetPath Local Target path for working or home dirctory for all Manifest Cloned/Returned/Generated
	TargetPath string `json:"target-path"`
	// SubPath is a path relative to TargetPath + Path where PrimaryRepository is cloned and contains
	// directories with ClusterType and Phase bundles, example:
	// Repositories[PrimaryRepositoryName].Url = 'https://github.com/airshipit/treasuremap'
	// SubPath = "manifests"
	// you would expect that at treasuremap/manifests you would have ephemeral/initinfra and
	// ephemera/target directories, containing kustomize.yaml.
	SubPath string `json:"sub-path"`
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

// RepoAuth struct describes method of authentication agaist given repository
type RepoAuth struct {
	// Type of authentication method to be used with given repository
	// supported types are "ssh-key", "ssh-pass", "http-basic"
	Type string `json:"type,omitempty"`
	//KeyPassword is a password decrypt ssh private key (used with ssh-key auth type)
	KeyPassword string `json:"key-pass,omitempty"`
	// KeyPath is path to private ssh key on disk (used with ssh-key auth type)
	KeyPath string `json:"ssh-key,omitempty"`
	//HTTPPassword is password for basic http authentication (used with http-basic auth type)
	HTTPPassword string `json:"http-pass,omitempty"`
	// SSHPassword is password for ssh password authentication (used with ssh-pass)
	SSHPassword string `json:"ssh-pass,omitempty"`
	// Username to authenticate against git remote (used with any type)
	Username string `json:"username,omitempty"`
}

// RepoCheckout container holds information how to checkout repository
// Each field is mutually exclusive
type RepoCheckout struct {
	// CommitHash is full hash of the commit that will be used to checkout
	CommitHash string `json:"commit-hash,omitempty"`
	// Branch is the branch name to checkout
	Branch string `json:"branch"`
	// Tag is the tag name to checkout
	Tag string `json:"tag"`
	// RemoteRef is not supported currently TODO
	// RemoteRef is used for remote checkouts such as gerrit change requests/github pull request
	// for example refs/changes/04/691202/5
	// TODO Add support for fetching remote refs
	RemoteRef string `json:"remote-ref"`
	// ForceCheckout is a boolean to indicate whether to use the `--force` option when checking out
	ForceCheckout bool `json:"force"`
}

// Holds the complex cluster name information
// Encapsulates the different operations around using it.
type ClusterComplexName struct {
	clusterName string
	clusterType string
}

// Bootstrap holds configurations for bootstrap steps
type Bootstrap struct {
	// Configuration parameters for container
	Container *Container `json:"container,omitempty"`
	// Configuration parameters for ISO builder
	Builder *Builder `json:"builder,omitempty"`
	// Configuration parameters for ephmeral node remote management
	RemoteDirect *RemoteDirect `json:"remoteDirect,omitempty"`
}

// Container parameters
type Container struct {
	// Container volume directory binding.
	Volume string `json:"volume,omitempty"`
	// ISO generator container image URL
	Image string `json:"image,omitempty"`
	// Container Runtime Interface driver
	ContainerRuntime string `json:"containerRuntime,omitempty"`
}

// Builder parameters
type Builder struct {
	// Cloud Init user-data file name placed to the container volume root
	UserDataFileName string `json:"userDataFileName,omitempty"`
	// Cloud Init network-config file name placed to the container volume root
	NetworkConfigFileName string `json:"networkConfigFileName,omitempty"`
	// File name for output metadata
	OutputMetadataFileName string `json:"outputMetadataFileName,omitempty"`
}

// RemoteDirect configuration options
type RemoteDirect struct {
	// RemoteType specifies type of epehemeral node managfement (e.g redfish,
	// smash e.t.c.)
	RemoteType string `json:"remoteType,omitempty"`
	// IsoURL specifies url to download ISO image for epehemeral node
	IsoURL string `json:"isoUrl,omitempty"`
}

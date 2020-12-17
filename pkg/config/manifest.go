/*
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
	"sigs.k8s.io/yaml"
)

// Manifest is a tuple of references to a Manifest (how do Identify, collect ,
// find the yaml manifests that airship uses to perform its operations)
type Manifest struct {
	// PhaseRepositoryName is a name of the repo, that contains site/<site-name> directory
	// and is a starting point for building document bundle
	PhaseRepositoryName string `json:"phaseRepositoryName"`
	// InventoryRepositoryName is a name of the repo contains inventory objects
	// to be used mostly with baremetal deployments
	// If not defined PhaseRepositoryName will be used to locate inventory
	InventoryRepositoryName string `json:"inventoryRepositoryName"`
	// ExtraRepositories is the map of extra repositories addressable by a name
	Repositories map[string]*Repository `json:"repositories,omitempty"`
	// TargetPath Local Target path for working or home directory for all Manifest Cloned/Returned/Generated
	TargetPath string `json:"targetPath"`
	// MetadataPath path to a metadata file relative to TargetPath
	MetadataPath string `json:"metadataPath"`
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

// PhaseMeta holds phase metadata
type PhaseMeta struct {
	// path is a kustomize entrypoint against which we will build bundle with phase objects
	Path string `json:"path,omitempty"`
	// docEntryPointPrefix is the path prefix for documentEntryPoint field in the phase config
	// If it is defined in the manifest metadata then it will be prepended
	// to the documentEntryPoint defined in the phase itself. So in this case the full path will be
	// targetPath + phaseRepoDir + docEntryPointPrefix + documentEntryPoint
	// E.g. let
	// targetPath (defined in airship config file) be /tmp
	// phaseRepoDir (this is the last part of the repo url given in the airship config file) be reponame
	// docEntryPointPrefix (defined in metadata) be foo/bar and
	// documentEntryPoint (defined in a phase) be baz/xyz
	// then the full path to the document bundle will be /tmp/reponame/foo/bar/baz/xyz
	// If docEntryPointPrefix is empty or not given at all, then the full path will be
	// targetPath + phaseRepoDir + documentEntryPoint (in our case /tmp/reponame/baz/xyz)
	DocEntryPointPrefix string `json:"docEntryPointPrefix,omitempty"`
}

// Manifest functions
func (m *Manifest) String() string {
	yamlData, err := yaml.Marshal(&m)
	if err != nil {
		return ""
	}
	return string(yamlData)
}

// GetTargetPath returns TargetPath field
func (m *Manifest) GetTargetPath() string {
	return m.TargetPath
}

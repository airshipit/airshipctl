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

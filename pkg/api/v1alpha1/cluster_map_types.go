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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// ClusterMap represents cluster defined for this manifest
type ClusterMap struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Keys in this map MUST correspond to context names in kubeconfigs provided
	Map map[string]*Cluster `json:"map,omitempty"`
}

// Cluster uniquely identifies a cluster and its parent cluster
type Cluster struct {
	// Parent is a key in ClusterMap.Map that identifies the name of the parent(management) cluster
	Parent string `json:"parent,omitempty"`
	// KubeconfigContext is the context in kubeconfig, default is equals to clusterMap key
	Sources []KubeconfigSource `json:"kubeconfigSources"`
}

// KubeconfigSource describes source of the kubeconfig
type KubeconfigSource struct {
	Type       KubeconfigSourceType       `json:"type"`
	FileSystem KubeconfigSourceFilesystem `json:"filesystem,omitempty"`
	Bundle     KubeconfigSourceBundle     `json:"bundle,omitempty"`
	ClusterAPI KubeconfigSourceClusterAPI `json:"clusterAPI,omitempty"`
}

// KubeconfigSourceType type of source
type KubeconfigSourceType string

const (
	// KubeconfigSourceTypeFilesystem is used when you want kubeconfig to be taken from local filesystem
	KubeconfigSourceTypeFilesystem KubeconfigSourceType = "filesystem"
	// KubeconfigSourceTypeBundle use config document bundle to get kubeconfig
	KubeconfigSourceTypeBundle KubeconfigSourceType = "bundle"
	// KubeconfigSourceTypeClusterAPI use ClusterAPI to get kubeconfig, parent cluster must be specified
	KubeconfigSourceTypeClusterAPI KubeconfigSourceType = "clusterAPI"
)

// KubeconfigSourceFilesystem get kubeconfig from filesystem path
type KubeconfigSourceFilesystem struct {
	Path    string `json:"path,omitempty"`
	Context string `json:"contextName,omitempty"`
}

// KubeconfigSourceClusterAPI get kubeconfig from clusterAPI parent cluster
type KubeconfigSourceClusterAPI struct {
	NamespacedName `json:"clusterNamespacedName,omitempty"`
}

// KubeconfigSourceBundle get kubeconfig from bundle
type KubeconfigSourceBundle struct {
	Context string `json:"contextName,omitempty"`
}

// NamespacedName is a name combined with namespace to uniquely identify objects
type NamespacedName struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// DefaultClusterMap can be used to safely unmarshal ClusterMap object without nil pointers
func DefaultClusterMap() *ClusterMap {
	return &ClusterMap{
		Map: make(map[string]*Cluster),
	}
}

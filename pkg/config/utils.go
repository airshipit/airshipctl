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

import "strings"

const (
	DefaultTestPrimaryRepo = "primary"
)

// NewConfig returns a newly initialized Config object
func NewConfig() *Config {
	return &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigAPIVersion,
		Clusters:   make(map[string]*ClusterPurpose),
		AuthInfos:  make(map[string]*AuthInfo),
		Contexts: map[string]*Context{
			AirshipDefaultContext: {
				Manifest: AirshipDefaultManifest,
			},
		},
		Manifests: map[string]*Manifest{
			AirshipDefaultManifest: {
				Repositories: map[string]*Repository{
					DefaultTestPrimaryRepo: {
						URLString: AirshipDefaultManifestRepoLocation,
						CheckoutOptions: &RepoCheckout{
							CommitHash: "master",
							Branch:     "master",
							RemoteRef:  "master",
						},
					},
				},
				TargetPath:            "/tmp/" + AirshipDefaultManifest,
				PrimaryRepositoryName: DefaultTestPrimaryRepo,
				SubPath:               AirshipDefaultManifestRepo + "/manifests/site",
			},
		},
		ModulesConfig: &Modules{
			BootstrapInfo: map[string]*Bootstrap{
				AirshipDefaultContext: {
					Container: &Container{
						Volume:           "/srv/iso:/config",
						Image:            AirshipDefaultBootstrapImage,
						ContainerRuntime: "docker",
					},
					Builder: &Builder{
						UserDataFileName:       "user-data",
						NetworkConfigFileName:  "network-config",
						OutputMetadataFileName: "output-metadata.yaml",
					},
					RemoteDirect: &RemoteDirect{
						RemoteType: AirshipDefaultRemoteType,
						IsoURL:     AirshipDefaultIsoURL,
					},
				},
			},
		},
	}
}

// NewContext is a convenience function that returns a new Context
func NewContext() *Context {
	return &Context{}
}

// NewCluster is a convenience function that returns a new Cluster
func NewCluster() *Cluster {
	return &Cluster{}
}

// NewManifest is a convenience function that returns a new Manifest
// object with non-nil maps
func NewManifest() *Manifest {
	return &Manifest{
		PrimaryRepositoryName: DefaultTestPrimaryRepo,
		Repositories:          map[string]*Repository{DefaultTestPrimaryRepo: NewRepository()},
	}
}

func NewRepository() *Repository {
	return &Repository{}
}

func NewAuthInfo() *AuthInfo {
	return &AuthInfo{}
}

func NewModules() *Modules {
	return &Modules{
		BootstrapInfo: make(map[string]*Bootstrap),
	}
}

// NewClusterPurpose is a convenience function that returns a new ClusterPurpose
func NewClusterPurpose() *ClusterPurpose {
	return &ClusterPurpose{
		ClusterTypes: make(map[string]*Cluster),
	}
}

// NewClusterComplexName returns a ClusterComplexName with the given name and type.
func NewClusterComplexName(clusterName, clusterType string) ClusterComplexName {
	return ClusterComplexName{
		Name: clusterName,
		Type: clusterType,
	}
}

// NewClusterComplexNameFromKubeClusterName takes the name of a cluster in a
// format which might be found in a kubeconfig file. This may be a simple
// string (e.g. myCluster), or it may be prepended with the type of the cluster
// (e.g. myCluster_target)
//
// If a valid cluster type was appended, the returned ClusterComplexName will
// have that type. If no cluster type is provided, the
// AirshipDefaultClusterType will be used.
func NewClusterComplexNameFromKubeClusterName(kubeClusterName string) ClusterComplexName {
	parts := strings.Split(kubeClusterName, AirshipClusterNameSeparator)

	if len(parts) == 1 {
		return NewClusterComplexName(kubeClusterName, AirshipDefaultClusterType)
	}

	// kubeClusterName matches the format myCluster_something.
	// Let's check if "something" is a clusterType.
	potentialType := parts[len(parts)-1]
	for _, ct := range AllClusterTypes {
		if potentialType == ct {
			// Rejoin the parts in the case of "my_cluster_etc_etc_<clusterType>"
			name := strings.Join(parts[:len(parts)-1], AirshipClusterNameSeparator)
			return NewClusterComplexName(name, potentialType)
		}
	}

	// "something" is not a valid clusterType, so just use the default
	return NewClusterComplexName(kubeClusterName, AirshipDefaultClusterType)
}

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
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

const (
	// DefaultTestPrimaryRepo holds default repo name
	DefaultTestPrimaryRepo = "primary"
)

// NewConfig returns a newly initialized Config object
func NewConfig() *Config {
	return &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigAPIVersion,
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
					IsoURL: AirshipDefaultIsoURL,
				},
			},
		},
		Clusters:  make(map[string]*ClusterPurpose),
		AuthInfos: make(map[string]*AuthInfo),
		Contexts: map[string]*Context{
			AirshipDefaultContext: {
				Manifest: AirshipDefaultManifest,
			},
		},
		ManagementConfiguration: map[string]*ManagementConfiguration{
			AirshipDefaultContext: {
				Type:                redfish.ClientType,
				Insecure:            true,
				UseProxy:            false,
				SystemActionRetries: DefaultSystemActionRetries,
				SystemRebootDelay:   DefaultSystemRebootDelay,
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

// NewRepository is a convenience function that returns a new Repository
func NewRepository() *Repository {
	return &Repository{}
}

// NewAuthInfo is a convenience function that returns a new AuthInfo
func NewAuthInfo() *AuthInfo {
	return &AuthInfo{}
}

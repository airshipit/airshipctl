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
	"encoding/base64"
)

// NewConfig returns a newly initialized Config object
func NewConfig() *Config {
	return &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigAPIVersion,
		BootstrapInfo: map[string]*Bootstrap{
			AirshipDefaultBootstrapInfo: {
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
		CurrentContext: AirshipDefaultContext,
		ManagementConfiguration: map[string]*ManagementConfiguration{
			AirshipDefaultManagementConfiguration: NewManagementConfiguration(),
		},
		Manifests: map[string]*Manifest{
			AirshipDefaultManifest: {
				Repositories: map[string]*Repository{
					DefaultTestPrimaryRepo: {
						URLString: AirshipDefaultManifestRepoLocation,
						CheckoutOptions: &RepoCheckout{
							Branch: "master",
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
	return &Cluster{
		NameInKubeconf:          "",
		Bootstrap:               AirshipDefaultBootstrapInfo,
		ManagementConfiguration: AirshipDefaultManagementConfiguration,
	}
}

// NewManifest is a convenience function that returns a new Manifest
// object with non-nil maps
func NewManifest() *Manifest {
	return &Manifest{
		PrimaryRepositoryName: DefaultTestPrimaryRepo,
		TargetPath:            DefaultTargetPath,
		SubPath:               DefaultSubPath,
		Repositories:          map[string]*Repository{DefaultTestPrimaryRepo: NewRepository()},
		MetadataPath:          DefaultManifestMetadataFile,
	}
}

// NewRepository is a convenience function that returns a new Repository
func NewRepository() *Repository {
	return &Repository{
		CheckoutOptions: &RepoCheckout{},
	}
}

// NewAuthInfo is a convenience function that returns a new AuthInfo
func NewAuthInfo() *AuthInfo {
	return &AuthInfo{}
}

// EncodeString returns the base64 encoding of given string
func EncodeString(given string) string {
	return base64.StdEncoding.EncodeToString([]byte(given))
}

// DecodeString returns the base64 decoded string
// If err decoding, return the given string
func DecodeString(given string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(given)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

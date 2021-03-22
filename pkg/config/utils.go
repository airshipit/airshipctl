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
	"encoding/base64"

	"opendev.org/airship/airshipctl/pkg/fs"
)

// NewConfig returns a newly initialized Config object
func NewConfig() *Config {
	return &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigAPIVersion,
		Permissions: Permissions{
			DirectoryPermission: AirshipDefaultDirectoryPermission,
			FilePermission:      AirshipDefaultFilePermission,
		},
		Contexts: map[string]*Context{
			AirshipDefaultContext: {
				Manifest:                AirshipDefaultManifest,
				ManagementConfiguration: AirshipDefaultManagementConfiguration,
			},
		},
		CurrentContext: AirshipDefaultContext,
		ManagementConfiguration: map[string]*ManagementConfiguration{
			AirshipDefaultManagementConfiguration: NewManagementConfiguration(),
		},
		Manifests: map[string]*Manifest{
			AirshipDefaultManifest: {
				Repositories: map[string]*Repository{
					DefaultTestPhaseRepo: {
						URLString: AirshipDefaultManifestRepoLocation,
						CheckoutOptions: &RepoCheckout{
							Branch: "master",
						},
					},
				},
				TargetPath:              "/tmp/" + AirshipDefaultManifest,
				PhaseRepositoryName:     DefaultTestPhaseRepo,
				InventoryRepositoryName: DefaultTestPhaseRepo,
				MetadataPath:            DefaultManifestMetadataFile,
			},
		},
		fileSystem: fs.NewDocumentFs(),
	}
}

// NewEmptyConfig returns an initialized Config object with no default values
func NewEmptyConfig() *Config {
	return &Config{
		ManagementConfiguration: map[string]*ManagementConfiguration{},
		Manifests:               map[string]*Manifest{},
		Contexts:                map[string]*Context{},
		fileSystem:              fs.NewDocumentFs(),
		Permissions: Permissions{
			DirectoryPermission: AirshipDefaultDirectoryPermission,
			FilePermission:      AirshipDefaultFilePermission,
		},
	}
}

// NewContext is a convenience function that returns a new Context
func NewContext() *Context {
	return &Context{}
}

// NewManifest is a convenience function that returns a new Manifest
// object with non-nil maps
func NewManifest() *Manifest {
	return &Manifest{
		InventoryRepositoryName: DefaultTestPhaseRepo,
		PhaseRepositoryName:     DefaultTestPhaseRepo,
		TargetPath:              DefaultTargetPath,
		Repositories:            map[string]*Repository{DefaultTestPhaseRepo: NewRepository()},
		MetadataPath:            DefaultManifestMetadataFile,
	}
}

// NewRepository is a convenience function that returns a new Repository
func NewRepository() *Repository {
	return &Repository{
		CheckoutOptions: &RepoCheckout{},
	}
}

// EncodeString returns the base64 encoding of given string
func EncodeString(given string) string {
	return base64.StdEncoding.EncodeToString([]byte(given))
}

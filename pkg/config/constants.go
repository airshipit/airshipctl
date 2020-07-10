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

package config

import "opendev.org/airship/airshipctl/pkg/remote/redfish"

// Constants related to the ClusterType type
const (
	Ephemeral                   = "ephemeral"
	Target                      = "target"
	AirshipClusterNameSeparator = "_"
	AirshipDefaultClusterType   = Target
)

// Constants related to Phases
const (
	InitinfraPhase  = "initinfra"
	ClusterctlPhase = InitinfraPhase
	BootstrapPhase  = "bootstrap"
)

// AllClusterTypes holds cluster types
var AllClusterTypes = [2]string{Ephemeral, Target}

// Constants defining default values
const (
	AirshipConfig                         = "config"
	AirshipConfigAPIVersion               = AirshipConfigGroup + "/" + AirshipConfigVersion
	AirshipConfigDir                      = ".airship"
	AirshipConfigEnv                      = "AIRSHIPCONFIG"
	AirshipConfigGroup                    = "airshipit.org"
	AirshipConfigKind                     = "Config"
	AirshipConfigVersion                  = "v1alpha1"
	AirshipDefaultBootstrapInfo           = "default"
	AirshipDefaultContext                 = "default"
	AirshipDefaultManagementConfiguration = "default"
	AirshipDefaultManifest                = "default"
	AirshipDefaultManifestRepo            = "treasuremap"
	AirshipDefaultManifestRepoLocation    = "https://opendev.org/airship/" + AirshipDefaultManifestRepo
	AirshipKubeConfig                     = "kubeconfig"
	AirshipKubeConfigEnv                  = "AIRSHIP_KUBECONFIG"
	AirshipPluginPath                     = "kustomize-plugins"
	AirshipPluginPathEnv                  = "AIRSHIP_KUSTOMIZE_PLUGINS"

	// Modules
	AirshipDefaultBootstrapImage = "quay.io/airshipit/isogen:latest-debian_stable"
	AirshipDefaultIsoURL         = "http://localhost:8099/debian-custom.iso"
	AirshipDefaultManagementType = redfish.ClientType
)

// Default values for remote operations
const (
	DefaultSystemActionRetries = 30
	DefaultSystemRebootDelay   = 30
)

// Default Value for manifest
const (
	// DefaultTestPrimaryRepo holds default repo name
	DefaultTestPrimaryRepo = "primary"
	// DefaultTargetPath holds default target path
	DefaultTargetPath = "/tmp/default"
	// DefaultSubPath holds default sub path
	DefaultSubPath = "manifest/default"
	// DefaultManifestMetadataFile default path to manifest metadata file
	DefaultManifestMetadataFile = "manifests/site/test-site/metadata.yaml"
)

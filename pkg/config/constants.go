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

// Constants related to Phases
const (
	Ephemeral       = "ephemeral"
	InitinfraPhase  = "initinfra"
	ClusterctlPhase = InitinfraPhase
	BootstrapPhase  = "bootstrap-iso"
)

// Constants defining default values
const (
	AirshipConfig                         = "config"
	AirshipConfigAPIVersion               = AirshipConfigGroup + "/" + AirshipConfigVersion
	AirshipConfigDir                      = ".airship"
	AirshipConfigEnv                      = "AIRSHIPCONFIG"
	AirshipConfigGroup                    = "airshipit.org"
	AirshipConfigKind                     = "Config"
	AirshipConfigVersion                  = "v1alpha1"
	AirshipDefaultContext                 = "default"
	AirshipDefaultDirectoryPermission     = 0750
	AirshipDefaultFilePermission          = 0640
	AirshipDefaultManagementConfiguration = "default"
	AirshipDefaultManifest                = "default"
	AirshipDefaultManifestRepo            = "treasuremap"
	AirshipDefaultManifestRepoLocation    = "https://opendev.org/airship/" + AirshipDefaultManifestRepo

	// Modules
	AirshipDefaultManagementType = redfish.ClientType

	//HomeEnvVar holds value of HOME directory from env
	HomeEnvVar = "$HOME"
)

// Default values for remote operations
const (
	DefaultSystemActionRetries = 30
	DefaultSystemRebootDelay   = 30
)

// Default Value for manifest
const (
	// DefaultTestPhaseRepo holds default repo name
	DefaultTestPhaseRepo = "primary"
	// DefaultTargetPath holds default target path
	DefaultTargetPath = "/tmp/default"
	// DefaultManifestMetadataFile default path to manifest metadata file
	DefaultManifestMetadataFile = "manifests/site/test-site/metadata.yaml"
)

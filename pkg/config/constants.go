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
	InitinfraPhase = "initinfra"
	BootstrapPhase = "bootstrap"
)

// AllClusterTypes holds cluster types
var AllClusterTypes = [2]string{Ephemeral, Target}

// Constants defining default values
const (
	AirshipConfigGroup                 = "airshipit.org"
	AirshipConfigVersion               = "v1alpha1"
	AirshipConfigAPIVersion            = AirshipConfigGroup + "/" + AirshipConfigVersion
	AirshipConfigKind                  = "Config"
	AirshipConfigDir                   = ".airship"
	AirshipConfig                      = "config"
	AirshipKubeConfig                  = "kubeconfig"
	AirshipConfigEnv                   = "AIRSHIPCONFIG"
	AirshipKubeConfigEnv               = "AIRSHIP_KUBECONFIG"
	AirshipDefaultContext              = "default"
	AirshipDefaultManifest             = "default"
	AirshipDefaultManifestRepo         = "treasuremap"
	AirshipDefaultManifestRepoLocation = "https://opendev.org/airship/" + AirshipDefaultManifestRepo

	// Modules
	AirshipDefaultBootstrapImage = "quay.io/airshipit/isogen:latest"
	AirshipDefaultIsoURL         = "http://localhost:8099/debian-custom.iso"
	AirshipDefaultRemoteType     = redfish.ClientType
)

const (
	FlagAPIServer    = "server"
	FlagAuthInfoName = "user"
	FlagBearerToken  = "token"
	FlagCAFile       = "certificate-authority"
	FlagCertFile     = "client-certificate"
	FlagClusterName  = "cluster"
	FlagClusterType  = "cluster-type"

	FlagCurrentContext = "current-context"
	FlagConfigFilePath = "airshipconf"
	FlagEmbedCerts     = "embed-certs"

	FlagInsecure  = "insecure-skip-tls-verify"
	FlagKeyFile   = "client-key"
	FlagManifest  = "manifest"
	FlagNamespace = "namespace"
	FlagPassword  = "password"

	FlagUsername = "username"
	FlagCurrent  = "current"
)

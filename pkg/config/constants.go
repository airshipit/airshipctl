package config

// Constants related to the ClusterType type
const (
	Ephemeral                   = "ephemeral"
	Target                      = "target"
	AirshipClusterNameSeparator = "_"
	AirshipDefaultClusterType   = Target
)

// Constants related to Phases
const (
	Initinfra = "initinfra"
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
	AirshipDefaultRemoteType     = "redfish"
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

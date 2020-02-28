package config

// OutputFormat denotes the form with which to display tabulated data
type OutputFormat string

// Constants related to the ClusterType type
const (
	Ephemeral                 = "ephemeral"
	Target                    = "target"
	AirshipClusterNameSep     = "_"
	AirshipClusterDefaultType = Target
)

// Sorted
var AllClusterTypes = [2]string{Ephemeral, Target}

// Constants defining default values
const (
	AirshipConfigGroup      = "airshipit.org"
	AirshipConfigVersion    = "v1alpha1"
	AirshipConfigAPIVersion = AirshipConfigGroup + "/" + AirshipConfigVersion
	AirshipConfigKind       = "Config"

	AirshipConfigDir  = ".airship"
	AirshipConfig     = "config"
	AirshipKubeConfig = "kubeconfig"

	AirshipConfigEnv     = "AIRSHIPCONFIG"
	AirshipKubeConfigEnv = "AIRSHIP_KUBECONFIG"

	AirshipDefaultContext              = "default"
	AirshipDefaultManifest             = "default"
	AirshipDefaultManifestRepo         = "treasuremap"
	AirshipDefaultManifestRepoLocation = "https://opendev.org/airship/" + AirshipDefaultManifestRepo

	// Modules
	AirshipDefaultBootstrapImage = "quay.io/airshipit/isogen:latest"
	AirshipDefaultIsoURL         = "http://localhost:8099/debian-custom.iso"
	AirshipDefaultRemoteType     = "redfish"
)

// Constants defining CLI flags
const (
	FlagAPIServer        = "server"
	FlagAuthInfoName     = "user"
	FlagBearerToken      = "token"
	FlagCAFile           = "certificate-authority"
	FlagCertFile         = "client-certificate"
	FlagClusterName      = "cluster"
	FlagClusterType      = "cluster-type"
	FlagContext          = "context"
	FlagCurrentContext   = "current-context"
	FlagConfigFilePath   = "airshipconf"
	FlagEmbedCerts       = "embed-certs"
	FlagImpersonate      = "as"
	FlagImpersonateGroup = "as-group"
	FlagInsecure         = "insecure-skip-tls-verify"
	FlagKeyFile          = "client-key"
	FlagManifest         = "manifest"
	FlagNamespace        = "namespace"
	FlagPassword         = "password"
	FlagTimeout          = "request-timeout"
	FlagUsername         = "username"
	FlagCurrent          = "current"
)

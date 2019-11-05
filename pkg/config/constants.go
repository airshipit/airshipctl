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

//Sorted
var AllClusterTypes = [2]string{Ephemeral, Target}

// Constants defining default values
const (
	AirshipConfigEnv        = "airshipconf"
	AirshipConfig           = "config"
	AirshipConfigDir        = ".airship"
	AirshipConfigKind       = "Config"
	AirshipConfigVersion    = "v1alpha1"
	AirshipConfigGroup      = "airshipit.org"
	AirshipConfigApiVersion = AirshipConfigGroup + "/" + AirshipConfigVersion
	AirshipKubeConfig       = "kubeconfig"
)

// Constants defining CLI flags
const (
	FlagClusterName      = "cluster"
	FlagClusterType      = "cluster-type"
	FlagAuthInfoName     = "user"
	FlagContext          = "context"
	FlagConfigFilePath   = AirshipConfigEnv
	FlagNamespace        = "namespace"
	FlagAPIServer        = "server"
	FlagInsecure         = "insecure-skip-tls-verify"
	FlagCertFile         = "client-certificate"
	FlagKeyFile          = "client-key"
	FlagCAFile           = "certificate-authority"
	FlagEmbedCerts       = "embed-certs"
	FlagBearerToken      = "token"
	FlagImpersonate      = "as"
	FlagImpersonateGroup = "as-group"
	FlagUsername         = "username"
	FlagPassword         = "password"
	FlagTimeout          = "request-timeout"
	FlagManifest         = "manifest"
)

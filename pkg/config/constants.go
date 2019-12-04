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
	FlagAPIServer        = "server"
	FlagAuthInfoName     = "user"
	FlagBearerToken      = "token"
	FlagCAFile           = "certificate-authority"
	FlagCertFile         = "client-certificate"
	FlagClusterName      = "cluster"
	FlagClusterType      = "cluster-type"
	FlagContext          = "context"
	FlagCurrentContext   = "current-context"
	FlagConfigFilePath   = AirshipConfigEnv
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
)

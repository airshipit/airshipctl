package document

// Label Selectors
const (
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"
	InitInfraSelector         = BaseAirshipSelector + "/phase = initinfra"

	// DeployToK8sSelector checks that deploy-k8s label is not equal to true or True (string)
	// Please note that by default every document in the manifest is to be deployed to kubernetes cluster.
	DeployToK8sSelector = "config.kubernetes.io/local-config notin (True, true)"
)

// Kinds
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"
)

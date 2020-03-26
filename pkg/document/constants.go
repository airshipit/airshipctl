package document

const (
	// Label Selectors
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"
	InitInfraSelector         = BaseAirshipSelector + "/phase = initinfra"

	// Annotation Selectors
	// Please note that by default every document in the manifest is to be deployed to kubernetes cluster.
	// so this selector simply checks that deploy-k8s label is not equal to true or True (string)
	DeployToK8sSelector = "config.kubernetes.io/local-config notin (True, true)"
)

// Kinds
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"
)

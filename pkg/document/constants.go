package document

const (
	// Selectors
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"

	// Labels
	DeployedByLabel = BaseAirshipSelector + "/deployed"

	// Identifiers (Static label values)
	InitinfraIdentifier = "initinfra"
)

// Kinds
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"
)

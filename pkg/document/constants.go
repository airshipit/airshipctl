package document

const (
	// Selectors
	BaseAirshipSelector       = "airshipit.org"
	EphemeralHostSelector     = BaseAirshipSelector + "/ephemeral-node in (True, true)"
	EphemeralUserDataSelector = BaseAirshipSelector + "/ephemeral-user-data in (True, true)"
	InitInfraSelector         = BaseAirshipSelector + "/phase = initinfra"
)

// Kinds
const (
	SecretKind        = "Secret"
	BareMetalHostKind = "BareMetalHost"
)

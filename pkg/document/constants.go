package document

const (
	// Selectors
	BaseAirshipSelector      = "airshipit.org"
	EphemeralClusterSelector = BaseAirshipSelector + "/ephemeral in (True, true)"

	// Labels
	DeployedByLabel = BaseAirshipSelector + "/deployed"

	// Identifiers (Static label values)
	InitinfraIdentifier = "initinfra"
)

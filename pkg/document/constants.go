package document

// Document labels and annotations
const (
	// Selectors
	BaseAirshipSelector      = "airshipit.org"
	EphemeralClusterSelector = BaseAirshipSelector + "/ephemeral in (True, true)"
	TargetClusterSelector    = BaseAirshipSelector + "/target in (True, true)"

	// Labels
	DeployedByLabel = BaseAirshipSelector + "/deployed"

	// Identifiers (Static label values)
	InitinfraIdentifier = "initinfra"
)

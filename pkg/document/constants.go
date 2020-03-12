package document

const (
	// Selectors
	BaseAirshipSelector = "airshipit.org"
	ControlNodeSelector = BaseAirshipSelector + "/node-role=control-plane"

	// Labels
	DeployedByLabel = BaseAirshipSelector + "/deployed"

	// Identifiers (Static label values)
	InitinfraIdentifier = "initinfra"
)

package document

// Document labels and annotations
const (
	BaseAirshipSelector      = "airshipit.org"
	EphemeralClusterSelector = BaseAirshipSelector + "/ephemeral in (True, true)"
	TargetClusterSelector    = BaseAirshipSelector + "/target in (True, true)"
)

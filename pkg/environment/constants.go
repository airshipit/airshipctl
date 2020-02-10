package environment

// OutputFormat denotes the form with which to display tabulated data
type OutputFormat string

// These are valid values for OutputFormat
const (
	Default  = ""
	JSON     = "json"
	YAML     = "yaml"
	NameOnly = "name"
	Wide     = "wide"
)

const HomeEnvVar = "$HOME"

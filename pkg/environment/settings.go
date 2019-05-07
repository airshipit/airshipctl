package environment

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipctl is being used
	// out-of-cluster
	KubeConfigFilePath string

	// Debug is used for verbose output
	Debug bool
}

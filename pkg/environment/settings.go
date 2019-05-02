package environment

// AirshipADMSettings is a container for all of the settings needed by airshipadm
type AirshipADMSettings struct {
	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipadm is being used
	// out-of-cluster
	KubeConfigFilePath string
}

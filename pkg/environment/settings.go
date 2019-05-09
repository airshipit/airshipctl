package environment

import (
	restclient "k8s.io/client-go/rest"
)

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// KubeConfigFilePath is the path to the kubernetes configuration file.
	// This flag is only needed when airshipctl is being used
	// out-of-cluster
	KubeConfigFilePath string

	// KubeConfig contains the configuration details needed for interacting
	// with the cluster
	KubeConfig *restclient.Config

	// Debug is used for verbose output
	Debug bool
}

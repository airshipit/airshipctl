package cluster

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
)

var (
	// ClusterUse subcommand string
	ClusterUse = "cluster"
)

// NewClusterCommand returns cobra command object of the airshipctl cluster and adds it's subcommands.
func NewClusterCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	clusterRootCmd := &cobra.Command{
		Use: ClusterUse,
		// TODO: (kkalynovskyi) Add more description when more subcommands are added
		Short: "Control Kubernetes cluster",
		Long:  "Interactions with Kubernetes cluster, such as get status, deploy initial infrastructure",
	}

	return clusterRootCmd
}

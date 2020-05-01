/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"github.com/spf13/cobra"

	clusterctlcmd "opendev.org/airship/airshipctl/pkg/clusterctl/cmd"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	moveLong = `
Move Cluster API objects, provider specific objects and all dependencies to the target cluster.

Note: The destination cluster MUST have the required provider components installed.
`

	moveExample = `
Move Cluster API objects, provider specific objects and all dependencies to the target cluster.

  airshipctl cluster move --target-context <context name>
`
)

// NewMoveCommand creates a command to move capi and bmo resources to the target cluster
func NewMoveCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	var toKubeconfigContext string
	moveCmd := &cobra.Command{
		Use:     "move",
		Short:   "Move Cluster API objects, provider specific objects and all dependencies to the target cluster",
		Long:    moveLong[1:],
		Example: moveExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := clusterctlcmd.NewCommand(rootSettings)
			if err != nil {
				return err
			}
			return command.Move(toKubeconfigContext)
		},
	}

	moveCmd.Flags().StringVar(&toKubeconfigContext, "target-context", "",
		"Context to be used within the kubeconfig file for the target cluster. If empty, current context will be used.")
	return moveCmd
}

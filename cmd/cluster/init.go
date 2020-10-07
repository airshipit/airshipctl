/*
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
	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	initLong = `
Initialize cluster-api providers based on airshipctl document set.
document set must contain document of Kind: Clusterctl in phase initinfra.
Path to initinfra phase is built based on airshipctl config
<manifest.target-path>/<subpath>/ephemeral/initinfra.
Clusterctl document example:
---
apiVersion: airshipit.org/v1alpha1
kind: Clusterctl
metadata:
  labels:
    airshipit.org/deploy-k8s: "false"
  name: clusterctl-v1
init-options:
  core-provider: "cluster-api:v0.3.3"
  bootstrap-providers:
    - "kubeadm:v0.3.3"
  infrastructure-providers:
    - "metal3:v0.3.1"
  control-plane-providers:
    - "kubeadm:v0.3.3"
providers:
  - name: "metal3"
    type: "InfrastructureProvider"
    versions:
      v0.3.1: manifests/function/capm3/v0.3.1
  - name: "kubeadm"
    type: "BootstrapProvider"
    versions:
      v0.3.3: manifests/function/cabpk/v0.3.3
  - name: "cluster-api"
    type: "CoreProvider"
    versions:
      v0.3.3: manifests/function/capi/v0.3.3
  - name: "kubeadm"
    type: "ControlPlaneProvider"
    versions:
      v0.3.3: manifests/function/cacpk/v0.3.3
`

	initExample = `
# Initialize clusterctl providers and components
airshipctl cluster init
`
)

// NewInitCommand creates a command to deploy cluster-api
func NewInitCommand(cfgFactory config.Factory) *cobra.Command {
	var kubeconfig string
	initCmd := &cobra.Command{
		Use:     "init",
		Short:   "Deploy cluster-api provider components",
		Long:    initLong,
		Example: initExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			command, err := clusterctlcmd.NewCommand(cfgFactory, kubeconfig)
			if err != nil {
				return err
			}
			return command.Init()
		},
	}

	initCmd.Flags().StringVar(
		&kubeconfig,
		"kubeconfig",
		"",
		"Path to kubeconfig associated with cluster being managed")

	return initCmd
}

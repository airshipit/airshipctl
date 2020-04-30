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

package phase

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/phase/apply"
)

const (
	applyLong = `
Apply specific phase to kubernetes cluster such as control-plane, workloads, initinfra
`
	applyExample = `
# Apply initinfra phase to a cluster
airshipctl phase apply initinfra
`
)

// NewApplyCommand creates a command to apply phase to k8s cluster.
func NewApplyCommand(rootSettings *environment.AirshipCTLSettings, factory client.Factory) *cobra.Command {
	i := apply.NewOptions(rootSettings)

	applyCmd := &cobra.Command{
		Use:     "apply PHASE_NAME",
		Short:   "Apply phase to a cluster",
		Long:    applyLong[1:],
		Args:    cobra.ExactArgs(1),
		Example: applyExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			i.PhaseName = args[0]
			client, err := factory(rootSettings)
			if err != nil {
				return err
			}
			i.Client = client

			return i.Run()
		},
	}
	addApplyFlags(i, applyCmd)
	return applyCmd
}

func addApplyFlags(i *apply.Options, cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.BoolVar(
		&i.DryRun,
		"dry-run",
		false,
		"don't deliver documents to the cluster, simulate the changes instead")

	flags.BoolVar(
		&i.Prune,
		"prune",
		false,
		`if set to true, command will delete all kubernetes resources that are not`+
			` defined in airship documents and have airshipit.org/deployed=apply label`)
}

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

	"opendev.org/airship/airshipctl/pkg/cluster/initinfra"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	// TODO add labels in description, when we have them designed
	initInfraLong = `
Deploy initial infrastructure to kubernetes cluster such as
metal3.io, argo, tiller and other manifest documents with appropriate labels
`
	initInfraExample = `
# Deploy infrastructure to a cluster
airshipctl cluster initinfra
`
)

// NewInitInfraCommand creates a command to deploy initial airship infrastructure.
func NewInitInfraCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	i := initinfra.NewInfra(rootSettings)
	initinfraCmd := &cobra.Command{
		Use:     "initinfra",
		Short:   "Deploy initinfra components to cluster",
		Long:    initInfraLong[1:],
		Example: initInfraExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return i.Run()
		},
	}
	addInitinfraFlags(i, initinfraCmd)
	return initinfraCmd
}

func addInitinfraFlags(i *initinfra.Infra, cmd *cobra.Command) {
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
			` defined in airship documents and have airshipit.org/deployed=initinfra label`)

	flags.StringVar(
		&i.ClusterType,
		"cluster-type",
		"ephemeral",
		`select cluster type to deploy initial infrastructure to;`+
			` currently only ephemeral is supported`)
}

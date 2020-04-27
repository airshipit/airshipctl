/*
Copyright 2016 The Kubernetes Authors.

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

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

const (
	setClusterLong = `
Create or modify a cluster in the airshipctl config files.

Since a cluster can be either "ephemeral" or "target", you must specify
cluster-type when managing clusters.
`

	setClusterExample = `
# Set the server field on the ephemeral exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=ephemeral \
  --server=https://1.2.3.4

# Embed certificate authority data for the target exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=target \
  --client-certificate-authority=$HOME/.airship/ca/kubernetes.ca.crt \
  --embed-certs

# Disable certificate checking for the target exampleCluster
airshipctl config set-cluster exampleCluster
  --cluster-type=target \
  --insecure-skip-tls-verify

# Configure client certificate for the target exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=target \
  --embed-certs \
  --client-certificate=$HOME/.airship/cert_file
`
)

// NewSetClusterCommand creates a command for creating and modifying clusters
// in the airshipctl config file.
func NewSetClusterCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.ClusterOptions{}
	cmd := &cobra.Command{
		Use:     "set-cluster NAME",
		Short:   "Manage clusters",
		Long:    setClusterLong[1:],
		Example: setClusterExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Name = args[0]
			modified, err := config.RunSetCluster(o, rootSettings.Config, true)
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %q of type %q modified.\n",
					o.Name, o.ClusterType)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %q of type %q created.\n",
					o.Name, o.ClusterType)
			}
			return nil
		},
	}

	addSetClusterFlags(o, cmd)
	return cmd
}

func addSetClusterFlags(o *config.ClusterOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.Server,
		"server",
		"",
		"server to use for the cluster")

	flags.StringVar(
		&o.ClusterType,
		"cluster-type",
		"",
		"the type of the cluster to add or modify")

	err := cmd.MarkFlagRequired("cluster-type")
	if err != nil {
		log.Fatal(err)
	}

	flags.BoolVar(
		&o.InsecureSkipTLSVerify,
		"insecure-skip-tls-verify",
		true,
		"if set, disable certificate checking")

	flags.StringVar(
		&o.CertificateAuthority,
		"certificate-authority",
		"",
		"path to a certificate authority")

	flags.BoolVar(
		&o.EmbedCAData,
		"embed-certs",
		false,
		"if set, embed the client certificate/key into the cluster")
}

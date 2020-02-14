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

var (
	setClusterLong = `
Sets a cluster entry in arshipctl config.
Specifying a name that already exists will merge new fields on top of existing values for those fields.`

	setClusterExample = fmt.Sprintf(`
# Set only the server field on the e2e cluster entry without touching other values.
airshipctl config set-cluster e2e --%v=ephemeral --%v=https://1.2.3.4

# Embed certificate authority data for the e2e cluster entry
airshipctl config set-cluster e2e --%v=target --%v-authority=~/.airship/e2e/kubernetes.ca.crt

# Disable cert checking for the dev cluster entry
airshipctl config set-cluster e2e --%v=target --%v=true

# Configure Client Certificate
airshipctl config set-cluster e2e --%v=target --%v=true --%v=".airship/cert_file"`,
		config.FlagClusterType,
		config.FlagAPIServer,
		config.FlagClusterType,
		config.FlagCAFile,
		config.FlagClusterType,
		config.FlagInsecure,
		config.FlagClusterType,
		config.FlagEmbedCerts,
		config.FlagCertFile)
)

// NewCmdConfigSetCluster creates a command object for the "set-cluster" action, which
// defines a new cluster airshipctl config.
func NewCmdConfigSetCluster(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	theCluster := &config.ClusterOptions{}

	setclustercmd := &cobra.Command{
		Use:     "set-cluster NAME",
		Short:   "Sets a cluster entry in the airshipctl config",
		Long:    setClusterLong,
		Example: setClusterExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			theCluster.Name = cmd.Flags().Args()[0]
			modified, err := config.RunSetCluster(theCluster, rootSettings.Config(), true)
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %q of type %q modified.\n",
					theCluster.Name, theCluster.ClusterType)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Cluster %q of type %q created.\n",
					theCluster.Name, theCluster.ClusterType)
			}
			return nil
		},
	}

	scInitFlags(theCluster, setclustercmd)
	return setclustercmd
}

func scInitFlags(o *config.ClusterOptions, setclustercmd *cobra.Command) {
	setclustercmd.Flags().StringVar(&o.Server, config.FlagAPIServer, o.Server,
		config.FlagAPIServer+" for the cluster entry in airshipctl config")

	setclustercmd.Flags().StringVar(&o.ClusterType, config.FlagClusterType, o.ClusterType,
		config.FlagClusterType+" for the cluster entry in airshipctl config")

	setclustercmd.Flags().BoolVar(&o.InsecureSkipTLSVerify, config.FlagInsecure, true,
		config.FlagInsecure+" for the cluster entry in airshipctl config")

	setclustercmd.Flags().StringVar(&o.CertificateAuthority, config.FlagCAFile, o.CertificateAuthority,
		"Path to "+config.FlagCAFile+" file for the cluster entry in airshipctl config")
	err := setclustercmd.MarkFlagFilename(config.FlagCAFile)
	if err != nil {
		log.Fatal(err)
	}

	setclustercmd.Flags().BoolVar(&o.EmbedCAData, config.FlagEmbedCerts, false,
		config.FlagEmbedCerts+" for the cluster entry in airshipctl config")
}

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
)

var (
	setAuthInfoLong = fmt.Sprintf(`Sets a user entry in airshipctl config
Specifying a name that already exists will merge new fields on top of existing values.`,
	)

	setAuthInfoExample = fmt.Sprintf(`
# Set only the "client-key" field on the "cluster-admin"
# entry, without touching other values:
airshipctl config set-credentials cluster-admin --%v=~/.kube/admin.key

# Set basic auth for the "cluster-admin" entry
airshipctl config set-credentials cluster-admin --%v=admin --%v=uXFGweU9l35qcif

# Embed client certificate data in the "cluster-admin" entry
airshipctl config set-credentials cluster-admin --%v=~/.kube/admin.crt --%v=true`,
		config.FlagUsername,
		config.FlagUsername,
		config.FlagPassword,
		config.FlagCertFile,
		config.FlagEmbedCerts,
	)
)

// NewCmdConfigSetAuthInfo creates a command object for the "set-credentials" action, which
// defines a new AuthInfo airship config.
func NewCmdConfigSetAuthInfo(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.AuthInfoOptions{}

	cmd := &cobra.Command{
		Use:     "set-credentials NAME",
		Short:   "Sets a user entry in the airshipctl config",
		Long:    setAuthInfoLong,
		Example: setAuthInfoExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Name = args[0]
			modified, err := config.RunSetAuthInfo(o, rootSettings.Config, true)
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "User information %q modified.\n", o.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "User information %q created.\n", o.Name)
			}
			return nil
		},
	}

	addSetAuthInfoFlags(o, cmd)
	return cmd
}

func addSetAuthInfoFlags(o *config.AuthInfoOptions, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVar(
		&o.ClientCertificate,
		config.FlagCertFile,
		"",
		"Path to "+config.FlagCertFile+" file for the user entry in airshipctl")

	flags.StringVar(
		&o.ClientKey,
		config.FlagKeyFile,
		"",
		"Path to "+config.FlagKeyFile+" file for the user entry in airshipctl")

	flags.StringVar(
		&o.Token,
		config.FlagBearerToken,
		"",
		config.FlagBearerToken+" for the user entry in airshipctl. Mutually exclusive with username and password flags.")

	flags.StringVar(
		&o.Username,
		config.FlagUsername,
		"",
		config.FlagUsername+" for the user entry in airshipctl. Mutually exclusive with token flag.")

	flags.StringVar(
		&o.Password,
		config.FlagPassword,
		"",
		config.FlagPassword+" for the user entry in airshipctl. Mutually exclusive with token flag.")

	flags.BoolVar(
		&o.EmbedCertData,
		config.FlagEmbedCerts,
		false,
		"Embed client cert/key for the user entry in airshipctl")
}

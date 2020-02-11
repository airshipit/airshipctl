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
	setAuthInfoLong = fmt.Sprintf(`Sets a user entry in airshipctl config
Specifying a name that already exists will merge new fields on top of existing values.

Client-certificate flags:
--%v=certfile --%v=keyfile

Bearer token flags:
--%v=bearer_token

Basic auth flags:
--%v=basic_user --%v=basic_password

Bearer token and basic auth are mutually exclusive.`,
		config.FlagCertFile,
		config.FlagKeyFile,
		config.FlagBearerToken,
		config.FlagUsername,
		config.FlagPassword)

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
	theAuthInfo := &config.AuthInfoOptions{}

	setauthinfo := &cobra.Command{
		Use:     "set-credentials NAME",
		Short:   "Sets a user entry in the airshipctl config",
		Long:    setAuthInfoLong,
		Example: setAuthInfoExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			theAuthInfo.Name = args[0]
			modified, err := config.RunSetAuthInfo(theAuthInfo, rootSettings.Config(), true)
			if err != nil {
				return err
			}
			if modified {
				fmt.Fprintf(cmd.OutOrStdout(), "User information %q modified.\n", theAuthInfo.Name)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "User information %q created.\n", theAuthInfo.Name)
			}
			return nil
		},
	}

	err := suInitFlags(theAuthInfo, setauthinfo)
	if err != nil {
		log.Fatal(err)
	}
	return setauthinfo
}

func suInitFlags(o *config.AuthInfoOptions, setauthinfo *cobra.Command) error {
	setauthinfo.Flags().StringVar(&o.ClientCertificate, config.FlagCertFile, o.ClientCertificate,
		"Path to "+config.FlagCertFile+" file for the user entry in airshipctl")
	err := setauthinfo.MarkFlagFilename(config.FlagCertFile)
	if err != nil {
		return err
	}

	setauthinfo.Flags().StringVar(&o.ClientKey, config.FlagKeyFile, o.ClientKey,
		"Path to "+config.FlagKeyFile+" file for the user entry in airshipctl")
	err = setauthinfo.MarkFlagFilename(config.FlagKeyFile)
	if err != nil {
		return err
	}

	setauthinfo.Flags().StringVar(&o.Token, config.FlagBearerToken, o.Token,
		config.FlagBearerToken+" for the user entry in airshipctl")

	setauthinfo.Flags().StringVar(&o.Username, config.FlagUsername, o.Username,
		config.FlagUsername+" for the user entry in airshipctl")

	setauthinfo.Flags().StringVar(&o.Password, config.FlagPassword, o.Password,
		config.FlagPassword+" for the user entry in airshipctl")

	setauthinfo.Flags().BoolVar(&o.EmbedCertData, config.FlagEmbedCerts, false,
		"Embed client cert/key for the user entry in airshipctl")

	return nil
}

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

const (
	setAuthInfoLong = `
Create or modify a user credential in the airshipctl config file.

Note that specifying more than one authentication method is an error.
`

	setAuthInfoExample = `
# Create a new user credential with basic auth
airshipctl config set-credentials exampleUser \
  --username=exampleUser \
  --password=examplePassword

# Change the client-key of a user named admin
airshipctl config set-credentials admin \
  --client-key=$HOME/.kube/admin.key

# Change the username and password of the admin user
airshipctl config set-credentials admin \
  --username=admin \
  --password=uXFGweU9l35qcif

# Embed client certificate data of the admin user
airshipctl config set-credentials admin \
  --client-certificate=$HOME/.kube/admin.crt \
  --embed-certs
`
)

// NewSetAuthInfoCommand creates a command for creating and modifying user
// credentials in the airshipctl config file.
func NewSetAuthInfoCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.AuthInfoOptions{}
	cmd := &cobra.Command{
		Use:     "set-credentials NAME",
		Short:   "Manage user credentials",
		Long:    setAuthInfoLong[1:],
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
		"client-certificate",
		"",
		"path to a certificate")

	flags.StringVar(
		&o.ClientKey,
		"client-key",
		"",
		"path to a key file")

	flags.StringVar(
		&o.Token,
		"token",
		"",
		"token to use for the credential; mutually exclusive with username and password flags.")

	flags.StringVar(
		&o.Username,
		"username",
		"",
		"username for the credential; mutually exclusive with token flag.")

	flags.StringVar(
		&o.Password,
		"password",
		"",
		"password for the credential; mutually exclusive with token flag.")

	flags.BoolVar(
		&o.EmbedCertData,
		"embed-certs",
		false,
		"if set, embed the client certificate/key into the credential")
}

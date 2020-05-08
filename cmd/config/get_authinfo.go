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

package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	getAuthInfoLong = `
Display a specific user's credentials, or all defined user
credentials if no name is provided.
`

	getAuthInfoExample = `
# List all user credentials
airshipctl config get-credentials

# Display a specific user's credentials
airshipctl config get-credentials exampleUser
`
)

// NewGetAuthInfoCommand creates a command for viewing the user credentials
// defined in the airshipctl config file.
func NewGetAuthInfoCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	o := &config.AuthInfoOptions{}
	cmd := &cobra.Command{
		Use:     "get-credentials [NAME]",
		Short:   "Get user credentials from the airshipctl config",
		Long:    getAuthInfoLong[1:],
		Example: getAuthInfoExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			airconfig := rootSettings.Config
			if len(args) == 1 {
				o.Name = args[0]
				authinfo, err := airconfig.GetAuthInfo(o.Name)
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), authinfo)
			} else {
				authinfos, err := airconfig.GetAuthInfos()
				if err != nil {
					return err
				}
				if len(authinfos) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No User credentials found in the configuration.")
				}
				for _, authinfo := range authinfos {
					fmt.Fprintln(cmd.OutOrStdout(), authinfo)
				}
			}
			return nil
		},
	}

	return cmd
}

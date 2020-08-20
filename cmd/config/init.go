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
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
)

const (
	initLong = `
Generate an airshipctl config file and its associated kubeConfig file.
These files will be written to the $HOME/.airship directory, and will contain
default configurations.

NOTE: This will overwrite any existing config files in $HOME/.airship
`
)

// NewInitCommand creates a command for generating default airshipctl config files.
func NewInitCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	// TODO(howell): It'd be nice to have a flag to tell
	// airshipctl where to store the new files.
	// TODO(howell): Currently, this command overwrites whatever the user
	// has in their airship directory. We should remove that functionality
	// as default and provide and optional --overwrite flag.
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate initial configuration files for airshipctl",
		Long:  initLong[1:],
		RunE: func(cmd *cobra.Command, args []string) error {
			return rootSettings.Config.PersistConfig(true)
		},
	}

	return cmd
}

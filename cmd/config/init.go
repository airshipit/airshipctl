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

package config

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
)

const (
	initLong = `
Generate an airshipctl config file. This file by default will be written to the $HOME/.airship directory,
and will contain default configuration. In case if flag --airshipconf provided - the file will be
written to the specified location instead. If a configuration file already exists at the specified path,
an error will be thrown; to overwrite it, specify the --overwrite flag.
`
	initExample = `
# Create new airshipctl config file at the default location
airshipctl config init

# Create new airshipctl config file at the custom location
airshipctl config init --airshipconf path/to/config

# Create new airshipctl config file at custom location and overwrite it
airshipctl config init --overwrite --airshipconf path/to/config
`
)

// NewInitCommand creates a command for generating default airshipctl config file.
func NewInitCommand() *cobra.Command {
	var overwrite bool
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Generate initial configuration file for airshipctl",
		Long:    initLong[1:],
		Example: initExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			airshipConfigPath, err := cmd.Flags().GetString("airshipconf")
			if err != nil {
				airshipConfigPath = ""
			}

			return config.CreateConfig(airshipConfigPath, overwrite)
		},
	}

	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite config file")
	return cmd
}

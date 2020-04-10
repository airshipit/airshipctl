/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmd

import (
	"io"

	"github.com/spf13/cobra"

	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"opendev.org/airship/airshipctl/cmd/baremetal"
	"opendev.org/airship/airshipctl/cmd/cluster"
	"opendev.org/airship/airshipctl/cmd/completion"
	"opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/cmd/document"
	"opendev.org/airship/airshipctl/cmd/secret"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// NewAirshipCTLCommand creates a root `airshipctl` command with the default commands attached
func NewAirshipCTLCommand(out io.Writer) (*cobra.Command, *environment.AirshipCTLSettings, error) {
	rootCmd, settings, err := NewRootCommand(out)
	return AddDefaultAirshipCTLCommands(rootCmd, settings), settings, err
}

// NewRootCommand creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCommand(out io.Writer) (*cobra.Command, *environment.AirshipCTLSettings, error) {
	settings := &environment.AirshipCTLSettings{}
	rootCmd := &cobra.Command{
		Use:           "airshipctl",
		Short:         "A unified entrypoint to various airship components",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(settings.Debug, cmd.OutOrStderr())

			// Load or Initialize airship Config
			settings.InitConfig()
		},
	}
	rootCmd.SetOut(out)
	rootCmd.AddCommand(NewVersionCommand())

	settings.InitFlags(rootCmd)

	return rootCmd, settings, nil
}

// AddDefaultAirshipCTLCommands is a convenience function for adding all of the
// default commands to airshipctl
func AddDefaultAirshipCTLCommands(cmd *cobra.Command, settings *environment.AirshipCTLSettings) *cobra.Command {
	cmd.AddCommand(baremetal.NewBaremetalCommand(settings))
	cmd.AddCommand(cluster.NewClusterCommand(settings))
	cmd.AddCommand(completion.NewCompletionCommand())
	cmd.AddCommand(document.NewDocumentCommand(settings))
	cmd.AddCommand(config.NewConfigCommand(settings))
	cmd.AddCommand(secret.NewSecretCommand())

	return cmd
}

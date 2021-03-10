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
	"path/filepath"

	"github.com/spf13/cobra"
	// Import to initialize client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"opendev.org/airship/airshipctl/cmd/baremetal"
	"opendev.org/airship/airshipctl/cmd/cluster"
	"opendev.org/airship/airshipctl/cmd/completion"
	"opendev.org/airship/airshipctl/cmd/config"
	"opendev.org/airship/airshipctl/cmd/document"
	"opendev.org/airship/airshipctl/cmd/phase"
	"opendev.org/airship/airshipctl/cmd/plan"
	"opendev.org/airship/airshipctl/cmd/secret"
	cfg "opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

// RootOptions stores global flags values
type RootOptions struct {
	Debug             bool
	AirshipConfigPath string
}

// NewAirshipCTLCommand creates a root `airshipctl` command with the default commands attached
func NewAirshipCTLCommand(out io.Writer) *cobra.Command {
	rootCmd, settings := NewRootCommand(out)
	return AddDefaultAirshipCTLCommands(rootCmd,
		cfg.CreateFactory(&settings.AirshipConfigPath))
}

// NewRootCommand creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCommand(out io.Writer) (*cobra.Command, *RootOptions) {
	options := &RootOptions{}
	rootCmd := &cobra.Command{
		Use:           "airshipctl",
		Short:         "A unified entrypoint to various airship components",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Init(options.Debug, cmd.ErrOrStderr())
		},
	}
	rootCmd.SetOut(out)
	initFlags(options, rootCmd)

	return rootCmd, options
}

// AddDefaultAirshipCTLCommands is a convenience function for adding all of the
// default commands to airshipctl
func AddDefaultAirshipCTLCommands(cmd *cobra.Command, factory cfg.Factory) *cobra.Command {
	cmd.AddCommand(baremetal.NewBaremetalCommand(factory))
	cmd.AddCommand(cluster.NewClusterCommand(factory))
	cmd.AddCommand(completion.NewCompletionCommand())
	cmd.AddCommand(document.NewDocumentCommand(factory))
	cmd.AddCommand(config.NewConfigCommand(factory))
	cmd.AddCommand(secret.NewSecretCommand(factory))
	cmd.AddCommand(phase.NewPhaseCommand(factory))
	cmd.AddCommand(plan.NewPlanCommand(factory))
	cmd.AddCommand(NewVersionCommand())

	return cmd
}

func initFlags(options *RootOptions, cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(&options.Debug, "debug", false, "enable verbose output")

	defaultAirshipConfigDir := filepath.Join(cfg.HomeEnvVar, cfg.AirshipConfigDir)

	defaultAirshipConfigPath := filepath.Join(defaultAirshipConfigDir, cfg.AirshipConfig)
	flags.StringVar(
		&options.AirshipConfigPath,
		"airshipconf",
		"",
		`Path to file for airshipctl configuration. (default "`+defaultAirshipConfigPath+`")`)
}

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

package baremetal

import (
	"time"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/inventory"
)

var (
	listHostsCommand = "list-hosts"
	listLong         = "List bare metal host(s)."
	listExample      = `
	Retrieve list of baremetal hosts, default output option is 'table'
	# airshipctl baremetal list-hosts
	# airshipctl baremetal list-hosts --namespace default
	# airshipctl baremetal list-hosts --namespace default --output table
	# airshipctl baremetal list-hosts --output yaml
`
)

// NewListHostsCommand provides a command to list a remote host.
func NewListHostsCommand(cfgFactory config.Factory, options *inventory.CommandOptions) *cobra.Command {
	l := &inventory.ListHostsCommand{Options: options}
	cmd := &cobra.Command{
		Use:     listHostsCommand,
		Short:   "Airshipctl command to list bare metal host(s)",
		Long:    listLong,
		Example: listExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			l.Writer = cmd.OutOrStdout()
			return l.RunE()
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&l.OutputFormat, "output", "o", "table", "output formats. Supported options are 'table' and 'yaml'")
	flags.StringVarP(&options.Namespace, flagNamespace, flagNamespaceSort, "", flagNamespaceDescription)
	flags.StringVarP(&options.Labels, flagLabel, flagLabelShort, "", flagLabelDescription)
	flags.DurationVar(&options.Timeout, flagTimeout, 10*time.Minute, flagTimeoutDescription)
	return cmd
}

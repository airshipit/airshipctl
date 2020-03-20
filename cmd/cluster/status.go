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

package cluster

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/cluster"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/k8s/client"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

// NewStatusCommand creates a command which reports the statuses of a cluster's deployed components.
func NewStatusCommand(rootSettings *environment.AirshipCTLSettings, factory client.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve statuses of deployed cluster components",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf := rootSettings.Config
			if err := conf.EnsureComplete(); err != nil {
				return err
			}

			manifest, err := conf.CurrentContextManifest()
			if err != nil {
				return err
			}

			docBundle, err := document.NewBundleByPath(manifest.TargetPath)
			if err != nil {
				return err
			}

			docs, err := docBundle.GetAllDocuments()
			if err != nil {
				return err
			}

			client, err := factory(rootSettings)
			if err != nil {
				return err
			}

			statusMap, err := cluster.NewStatusMap(client)
			if err != nil {
				return err
			}

			tw := util.NewTabWriter(cmd.OutOrStdout())
			fmt.Fprintf(tw, "Kind\tName\tStatus\n")
			for _, doc := range docs {
				status, err := statusMap.GetStatusForResource(doc)
				if err != nil {
					log.Debug(err)
				} else {
					fmt.Fprintf(tw, "%s\t%s\t%s\n", doc.GetKind(), doc.GetName(), status)
				}
			}
			tw.Flush()
			return nil
		},
	}

	return cmd
}

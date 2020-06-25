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

package document

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/log"
)

// NewDocumentCommand creates a new command for managing airshipctl documents
func NewDocumentCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	documentRootCmd := &cobra.Command{
		Use:   "document",
		Short: "Manage deployment documents",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Note: config is not loaded here; the kustomize plugin command doesn't
			// require it, and multiple use cases fail if we assume the file is there.
			log.Init(rootSettings.Debug, cmd.OutOrStderr())
		},
	}

	documentRootCmd.AddCommand(NewPullCommand(rootSettings, true))
	documentRootCmd.AddCommand(NewPluginCommand(rootSettings))

	return documentRootCmd
}

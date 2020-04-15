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

	"opendev.org/airship/airshipctl/pkg/document/render"
	"opendev.org/airship/airshipctl/pkg/environment"
	"opendev.org/airship/airshipctl/pkg/errors"
)

// NewRenderCommand create a new command for document rendering
func NewRenderCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	renderSettings := &render.Settings{AirshipCTLSettings: rootSettings}
	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render documents from model",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.ErrNotImplemented{}
		},
	}

	addRenderFlags(renderSettings, renderCmd)
	return renderCmd
}

// addRenderFlags adds flags for document render sub-command
func addRenderFlags(settings *render.Settings, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringArrayVarP(
		&settings.Label,
		"label",
		"l",
		nil,
		"filter documents by Labels")

	flags.StringArrayVarP(
		&settings.Annotation,
		"annotation",
		"a",
		nil,
		"filter documents by Annotations")

	flags.StringArrayVarP(
		&settings.GroupVersion,
		"apiversion",
		"g",
		nil,
		"filter documents by API version")

	flags.StringArrayVarP(
		&settings.Kind,
		"kind",
		"k",
		nil,
		"filter documents by Kinds")

	flags.StringVarP(
		&settings.RawFilter,
		"filter",
		"f",
		"",
		"logical expression for document filtering")
}

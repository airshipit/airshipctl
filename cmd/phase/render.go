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

package phase

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/phase"
)

const (
	renderLong = `
Render documents for a phase.
`

	renderExample = `
Get all 'initinfra' phase documents containing labels "app=helm" and "service=tiller"
# airshipctl phase render initinfra -l app=helm,service=tiller

Get all phase documents containing labels "app=helm" and "service=tiller" and kind 'Deployment'
# airshipctl phase render initinfra -l app=helm,service=tiller -k Deployment

Get all documents from config bundle
# airshipctl phase render --source config

Get all documents executor rendered documents for a phase
# airshipctl phase render initinfra --source executor
`
)

// NewRenderCommand create a new command for document rendering
func NewRenderCommand(cfgFactory config.Factory) *cobra.Command {
	filterOptions := &phase.RenderCommand{}
	renderCmd := &cobra.Command{
		Use:     "render PHASE_NAME",
		Short:   "Airshipctl command to render phase documents from model",
		Long:    renderLong[1:],
		Example: renderExample,
		Args:    RenderArgs(filterOptions),
		RunE: func(cmd *cobra.Command, args []string) error {
			return filterOptions.RunE(cfgFactory, cmd.OutOrStdout())
		},
	}

	addRenderFlags(filterOptions, renderCmd)
	return renderCmd
}

// addRenderFlags adds flags for document render sub-command
func addRenderFlags(filterOptions *phase.RenderCommand, cmd *cobra.Command) {
	flags := cmd.Flags()

	flags.StringVarP(&filterOptions.Label, "label", "l", "", "filter documents by Labels")
	flags.StringVarP(&filterOptions.Annotation, "annotation", "a", "", "filter documents by Annotations")
	flags.StringVarP(&filterOptions.APIVersion, "apiversion", "g", "", "filter documents by API version")
	flags.StringVarP(&filterOptions.Kind, "kind", "k", "", "filter documents by Kind")
	flags.StringVarP(&filterOptions.Source, "source", "s", phase.RenderSourcePhase,
		"phase: phase entrypoint will be rendered by kustomize, if entrypoint is not specified error will be returned\n"+
			"executor: rendering will be performed by executor if the phase\n"+
			"config: this will render bundle containing phase and executor documents")
	flags.BoolVarP(&filterOptions.FailOnDecryptionError, "decrypt", "d", false,
		"ensure that decryption of encrypted documents has finished successfully")
}

// RenderArgs returns an error if there are not exactly n args.
func RenderArgs(opts *phase.RenderCommand) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return &ErrRenderTooManyArgs{Count: len(args)}
		}
		if len(args) == 1 {
			opts.PhaseID.Name = args[0]
		}
		return nil
	}
}

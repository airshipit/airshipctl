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
		"Filter documents by Labels")

	flags.StringArrayVarP(
		&settings.Annotation,
		"annotation",
		"a",
		nil,
		"Filter documents by Annotations")

	flags.StringArrayVarP(
		&settings.GroupVersion,
		"apiversion",
		"g",
		nil,
		"Filter documents by API version")

	flags.StringArrayVarP(
		&settings.Kind,
		"kind",
		"k",
		nil,
		"Filter documents by Kinds")

	flags.StringVarP(
		&settings.RawFilter,
		"filter",
		"f",
		"",
		"Logical expression for document filtering")
}

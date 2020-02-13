package document

import (
	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/document/pull"
	"opendev.org/airship/airshipctl/pkg/environment"
)

// NewDocumentPullCommand creates a new command for pulling airship document repositories
func NewDocumentPullCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {
	settings := pull.Settings{AirshipCTLSettings: rootSettings}
	documentPullCmd := &cobra.Command{
		Use:   "pull",
		Short: "pulls documents from remote git repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return settings.Pull()
		},
	}

	return documentPullCmd
}

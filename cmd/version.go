package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"opendev.org/airship/airshipctl/pkg/util"
)

// NewVersionCommand prints out the versions of airshipctl and its underlying tools
func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipctl",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			clientV := clientVersion()
			w := util.NewTabWriter(out)
			defer w.Flush()
			fmt.Fprintf(w, "%s:\t%s\n", "airshipctl", clientV)
		},
	}
	return versionCmd
}

func clientVersion() string {
	// TODO(howell): There's gotta be a smarter way to do this
	return "v0.1.0"
}

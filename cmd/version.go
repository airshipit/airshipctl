package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// NewVersionCommand prints out the versions of airshipctl and its underlying tools
func NewVersionCommand(out io.Writer) *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipctl",
		Run: func(cmd *cobra.Command, args []string) {
			clientV := clientVersion()
			w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
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

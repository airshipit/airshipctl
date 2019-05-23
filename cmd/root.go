package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root `airshipctl` command. All other commands are
// subcommands branching from this one
func NewRootCmd(out io.Writer) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "airshipctl",
		Short: "airshipctl is a unified entrypoint to various airship components",
	}
	rootCmd.SetOutput(out)
	rootCmd.AddCommand(NewVersionCommand(out))
	return rootCmd, nil
}

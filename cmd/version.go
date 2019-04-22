package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of airshipadm",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("airshipadm v0.1.0")
		},
	}
	return versionCmd
}

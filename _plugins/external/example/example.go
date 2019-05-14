package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

//nolint:deadcode,unused,unparam
func NewCommand(out io.Writer, args []string) *cobra.Command {
	exampleCommand := &cobra.Command{
		Use:   "example",
		Short: "an example command",
		// Hidden is set to true because this is an example
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(out, "Hello world!")
		},
	}
	return exampleCommand
}

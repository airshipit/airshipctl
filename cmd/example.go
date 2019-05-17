package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func NewExampleCommand(out io.Writer, args []string) *cobra.Command {
	exampleCommand := &cobra.Command{
		Use:   "example",
		Short: "an example command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(out, "Hello world!")
		},
	}
	return exampleCommand
}

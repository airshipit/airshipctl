package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewWorkflowVersionCommand() *cobra.Command {
	workflowVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show the version number of argo",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("workflow vx.x.x")
		},
	}

	return workflowVersionCmd
}

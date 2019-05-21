package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/ian-howell/airshipctl/cmd/workflow"

	// "github.com/ian-howell/exampleplugin"
)

// pluginCommands are the functions that create the entrypoint command for a plugin
var pluginCommands = []func(io.Writer, []string) *cobra.Command{
	// exampleplugin.NewExampleCommand, // This is an example and shouldn't be enabled in production builds
	workflow.NewWorkflowCommand,
}

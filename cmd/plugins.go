package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

// builtinPlugins are the plugins that are built and maintained by the
// airshipctl team. They may be disabled if desired
var builtinPlugins = []func(io.Writer, []string) *cobra.Command{
	NewWorkflowCommand,
}

// externalPlugins are external. The function to create a command should be
// placed here
var externalPlugins = []func(io.Writer, []string) *cobra.Command{
	NewExampleCommand, // This is an example and shouldn't be enabled in production builds
}

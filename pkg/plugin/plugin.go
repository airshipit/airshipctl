package plugin

import (
	"fmt"
	"io"
	"plugin"

	"github.com/spf13/cobra"
)

const badInterfaceFormat = `plugin at %s is missing required function:
	- NewCommand(func io.writer, []string) *cobra.Command)`

func CreateCommandFromPlugin(pluginPath string, out io.Writer, args []string) (*cobra.Command, error) {
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		return nil, fmt.Errorf("plugin at %s could not be opened", pluginPath)
	}
	cmdSym, err := plug.Lookup("NewCommand")
	if err != nil {
		return nil, fmt.Errorf(badInterfaceFormat, pluginPath)
	}
	command, ok := cmdSym.(func(io.Writer, []string) *cobra.Command)
	if !ok {
		return nil, fmt.Errorf(badInterfaceFormat, pluginPath)
	}
	return command(out, args), nil
}

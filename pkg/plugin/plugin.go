package plugin

import (
	"io"
	"plugin"

	"github.com/spf13/cobra"
)

func CreateCommandFromPlugin(pluginPath string, out io.Writer, args []string) *cobra.Command {
	//TODO(howell): Remove these panics
	plug, err := plugin.Open(pluginPath)
	if err != nil {
		panic(err.Error())
	}
	cmdSym, err := plug.Lookup("NewCommand")
	if err != nil {
		panic(err.Error())
	}
	command, ok := cmdSym.(func(io.Writer, []string) *cobra.Command)
	if !ok {
		panic("NewCommand does not meet the interface.")
	}
	return command(out, args)
}

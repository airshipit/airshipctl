# airshipctl

### Custom Plugins Tutorial

This tutorial walks through a very basic plugin for `airshipctl`. For a more
involved example, see [Plugin Support](docs/plugins.md)

The following steps will get you started with a very rudimentary example plugin
for airshipctl. First, create a directory for your project outside of the
GOPATH:
```
mkdir /tmp/example
cd /tmp/example
```
This project will need to be a go module. You can initialize a module named
`example` with the following:
```
go mod init example
```
Note that modules are a relatively new feature added to Go, so you'll need to
be running Go1.11 or greater. Also note that most modules will follow a naming
schema that matches the remote version control system. A more realistice module
name might look something like `github.com/ian-howell/exampleplugin`.

Next, create a file `main.go` and populate it with the following:
```go
package main

import (
	"fmt"
	"os"

	"github.com/ian-howell/airshipctl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd, _, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create root airshipctl command: %s\n", err.Error())
		os.Exit(1)
	}

	exampleCmd := &cobra.Command{
		Use:   "example",
		Short: "an example plugin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(os.Stdout, "Hello airshipctl!")
		},
	}

	rootCmd.AddCommand(exampleCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failure during execution: %s\n", err.Error())
		os.Exit(1)
	}
}
```
And finally, run the build command to download and compile `airshipctl`:
```
go build -o airshipctl
```
Now that you've built `airshipctl`, you can access your plugin with the following command:
```
./airshipctl example
```

You may have noticed that this example ignores the second return value from
cmd.NewRootCmd. This value is a pointer to the AirshipCTLSettings, which
contains various configuration details, such as the debug flag and the path to
the config file\*. A useful paradigm involves embedding this object into a
custom ExampleSettings struct. This can be seen in the demo repo.

For a more involved example, see [Plugin Support](docs/plugins.md)

\* Work in progress

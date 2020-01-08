# Plugin Support

##### Table of Contents
* [Compile-In Plugins](#compile-in)
* [Fine Tuning a Build](#fine-tuning)
  * [Command Selection](#command-selection)
  * [Accessing `airshipctl` settings](#settings)

Our requirements for `airshipctl` contain two very conflicting concepts. One,
we'd like to assert that `airshipctl` is a statically linked executable, such
that it can be easily distributed. Two, we'd like to have plugin support. These
requirements can't coincide within the same project under the standard
definition of a plugin. Our solution is to provide a more refined definition of
what a plugin actually is.

<a name="compile-in" />

## Compile-In Plugins

In order to support plugins to an independent binary file, we use the concept
of *compile-in plugins*. A *compile-in plugin* is an add-on that is built into
the main application at compile time, as opposed to runtime. This means that
while `airshipctl` is a standalone application, it also acts as a sort of
library.  In fact, taking a deeper look at `airshipctl` reveals that the base
application is incredibly simple. At its core, `airshipctl` provides exactly 2
commands: `version` and `help`. Take a look at the following snippet to see
what this looks like:

```go
package main

import (
	"os"
	"opendev.org/airship/airshipctl/cmd"
)

func main() {
	rootCmd, _, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		panic(err)
	}
	rootCmd.Execute()
}
```

Compiling and running the above gives the following output:

```
$ ./airshipctl
airshipctl is a unified entrypoint to various airship components

Usage:
  airshipctl [command]

Available Commands:
  help        Help about any command
  version     Show the version number of airshipctl

Flags:
      --debug   enable verbose output
  -h, --help    help for airshipctl

Use "airshipctl [command] --help" for more information about a command.
```

Every other command is treated as a plugin. Changing `main` to the following
adds the default commands, or "plugins", to the`airshipctl` tool:

```go
func main() {
	rootCmd, settings, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		panic(err)
	}
	cmd.AddDefaultAirshipCTLCommands(rootCmd, settings)
	rootCmd.Execute()
}
```

Compiling and running now provides the following commands:

```
Available Commands:
  bootstrap   bootstraps airshipctl
  help        Help about any command
  version     Show the version number of airshipctl
  ------ more commands TBD ------
```

Downloading and building the main `airshipctl` project will default to
providing the builtin commands (such as `bootstrap`), much like the above. A
plugin author wishing to use `airshipctl` can then use the `rootCmd` as the
first of a series of building blocks. The following demonstrates the addition
of a new command, `hello`:

```go
package main

import (
	"fmt"
	"os"
	"opendev.org/airship/airshipctl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd, settings, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		panic(err)
	}

	cmd.AddDefaultAirshipCTLCommands(rootCmd, settings)

	helloCmd := &cobra.Command{
		Use: "hello",
		Short: "Prints a friendly message to the screen",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World!")
		},
	}
	rootCmd.AddCommand(helloCmd)

	rootCmd.Execute()
}
```

<a name="fine-tuning" />

## Fine Tuning a Build

There are a couple of ways in which a plugin author can fine tune their version
of `airshipctl`. These manifest as an ability to pick and choose various
plugins (including the defaults), and capabilities for accessing the same
settings as other `airshipctl` commands.

<a name="command-selection" />

### Command Selection

In the previous section, we introduced the `AddDefaultAirshipCTLCommands`
function. That command will simply dump all of the builtin commands onto the
root. But a plugin author might not need all of the builtins. To deal with
this, the author can pick and choose specific commands to add to their
`airshipctl`, much like the following:

```go
package main

import (
	"os"
	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/cmd/bootstrap"
)

func main() {
	rootCmd, settings, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(bootstrap.NewBootstrapCommand(settings))
	rootCmd.Execute()
}
```

This variant of `airshipctl` will have the `bootstrap` command, but will not
have any other builtins.

This can be particularly useful if a plugin author desires to "override" a
specific functionality provided by a builtin command. For example, you might
write your own `bootstrap` command and use it in place of the builtin.

<a name="settings" />

### Accessing `airshipctl` settings

The `airshipctl` will contain several settings which may be useful to a plugin
author. The following snippet demonstrates how to use the `debug` flag,
provided by `airshipctl`, as well as a custom `alt-message` flag, provided by
the plugin.

```go
package main

import (
	"fmt"
	"os"
	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/pkg/environment"
	"github.com/spf13/cobra"
)

type Settings struct {
	*environment.AirshipCTLSettings

	AltMessage bool
}

func main() {
	rootCmd, rootSettings, err := cmd.NewRootCmd(os.Stdout)
	if err != nil {
		panic(err)
	}

	settings := Settings{AirshipCTLSettings: rootSettings}
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Prints a friendly message to the screen",
		Run: func(cmd *cobra.Command, args []string) {
			if settings.Debug {
				fmt.Println("DEBUG: a debugging message")
			}
			if settings.AltMessage {
				fmt.Println("Goodbye World!")
			} else {
				fmt.Println("Hello World!")
			}
		},
	}
	helloCmd.PersistentFlags().BoolVar(&settings.AltMessage, "alt-message", false, "display an alternate message")
	rootCmd.AddCommand(helloCmd)

	rootCmd.Execute()
}
```

The `AirshipCTLSettings` object can be found
[here](../../pkg/environment/settings.go). Future documentation TBD.

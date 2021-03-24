# Plugin Support

## Table of Contents

* [Compile-In Plugins](#compile-in-plugins)
  * [Fine Tuning a Build](#fine-tuning-a-build)
    * [Command Selection](#command-selection)
    * [Accessing `airshipctl` options](#accessing-airshipctl-options)
* [Container Plugins](#container-plugins)

Our requirements for `airshipctl` contain two very conflicting concepts. One,
we'd like to assert that `airshipctl` is a statically linked executable, such
that it can be easily distributed. Two, we'd like to have plugin support. These
requirements can't coincide within the same project under the standard
definition of a plugin. Our solution is to provide a more refined definition of
what a plugin actually is.

## Compile-In Plugins

In order to support plugins to an independent binary file, we use the concept
of "compile-in plugins". A "compile-in plugin" is an add-on that is built into
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
	rootCmd, _ := cmd.NewRootCommand(os.Stdout)
	rootCmd.Execute()
}
```

Compiling and running the above gives the following output:

```
$ ./airshipctl
A unified entrypoint to various airship components
```

Every other command is treated as a plugin. Changing `main` to the following
adds the default commands, or "plugins", to the `airshipctl` tool:

```go
func main() {
	rootCmd, settings := cmd.NewRootCommand(os.Stdout)
	cmd.AddDefaultAirshipCTLCommands(rootCmd, cfg.CreateFactory(&settings.AirshipConfigPath))
	rootCmd.Execute()
}
```

Compiling and running now provides the following output:

```
$ ./airshipctl
A unified entrypoint to various airship components

Usage:
  airshipctl [command]

Available Commands:
  baremetal   Perform actions on baremetal hosts
  cluster     Manage Kubernetes clusters
  completion  Generate completion script for the specified shell (bash or zsh)
  config      Manage the airshipctl config file
  document    Manage deployment documents
  help        Help about any command
  image       Manage ISO image creation
  phase       Manage phases
  secret      Manage secrets
  version     Show the version number of airshipctl

Flags:
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
  -h, --help                 help for airshipctl

Use "airshipctl [command] --help" for more information about a command.
```

Downloading and building the main `airshipctl` project will default to
providing the builtin commands (such as `phase`), much like the above. A
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
	cfg "opendev.org/airship/airshipctl/pkg/config"
)

func main() {
	rootCmd, settings := cmd.NewRootCommand(os.Stdout)
	cmd.AddDefaultAirshipCTLCommands(rootCmd, cfg.CreateFactory(&settings.AirshipConfigPath))

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

### Fine Tuning a Build

There are a couple of ways in which a plugin author can fine tune their version
of `airshipctl`. These manifest as an ability to pick and choose various
plugins (including the defaults), and capabilities for accessing the same
settings as other `airshipctl` commands.

#### Command Selection

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
	"opendev.org/airship/airshipctl/cmd/phase"
	cfg "opendev.org/airship/airshipctl/pkg/config"
)

func main() {
	rootCmd, settings := cmd.NewRootCommand(os.Stdout)
	rootCmd.AddCommand(phase.NewPhaseCommand(cfg.CreateFactory(&settings.AirshipConfigPath)))
	rootCmd.Execute()
}
```

This variant of `airshipctl` will have the `phase` command, but will not
have any other builtins.

This can be particularly useful if a plugin author desires to "override" a
specific functionality provided by a builtin command. For example, you might
write your own `phase` command and use it in place of the builtin.

#### Accessing `airshipctl` options

A plugin author can define plugin options and/or use root command options.
The following snippet demonstrates how to use the `debug` flag,
provided by root command options, as well as a custom `alt-message` flag, provided by
the plugin.

```go
package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"opendev.org/airship/airshipctl/cmd"
	"os"
)

type Options struct {
	*cmd.RootOptions
	AltMessage bool
}

func main() {
	rootCmd, rootOptions := cmd.NewRootCommand(os.Stdout)
	options := &Options{RootOptions: rootOptions}
	helloCmd := &cobra.Command{
		Use:   "hello",
		Short: "Prints a friendly message to the screen",
		Run: func(cmd *cobra.Command, args []string) {
			if options.Debug {
				fmt.Println("Debug message")
			}
			if !options.AltMessage {
				fmt.Println("Hello World!")
			} else {
				fmt.Println("Goodbye World!")
			}
		},
	}

	helloCmd.PersistentFlags().BoolVar(&options.AltMessage, "alt-message", false, "display an alternate message")
	rootCmd.AddCommand(helloCmd)
	rootCmd.Execute()
}
```

## Container Plugins

`airshipctl` is mostly focused on managing Kubernetes cluster lifecycle using yaml documents. `airshipctl` uses
`kustomize` capabilities to deal with bundles of yaml documents. In turn, `kustomize` provides a way to
generate/transform yaml documents using plugins (functions). We can define a yaml document with the annotation
as follows

```yaml
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/airshipit/templater:v2
values:
  hosts:
  - macAddress: 00:aa:bb:cc:dd
    name: node-1
  - macAddress: 00:aa:bb:cc:ee
    name: node-2
template: |
  {{ range .hosts -}}
  ---
  apiVersion: metal3.io/v1alpha1
  kind: BareMetalHost
  metadata:
    name: {{ .name }}
  spec:
    bootMACAddress: {{ .macAddress }}
  {{ end -}}
```

`kustomize` looks at the annotation `config.kubernetes.io/function` and runs the container with the image defined in the
annotation. The container usually accepts a bunch of yaml documents on its stdin and
outputs a generated/modified bunch of yaml documents on its output. The document in the above example defines the
configuration for the template plugin. This particular example generates two `BareMetalHost` documents.

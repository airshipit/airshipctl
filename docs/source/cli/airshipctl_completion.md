## airshipctl completion

Airshipctl command to generate completion script for the specified shell (bash or zsh)

### Synopsis

Generate completion script for airshipctl for the specified shell (bash or zsh).


```
airshipctl completion SHELL [flags]
```

### Examples

```

Save shell completion to a file
# airshipctl completion bash > $HOME/.airship_completions

Apply completions to the current shell
# source <(airshipctl completion bash)

```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl](airshipctl.md)	 - A unified command line tool for management of end-to-end kubernetes cluster deployment on cloud infrastructure environments.


## airshipctl completion

Generate completion script for the specified shell (bash or zsh)

### Synopsis

Generate completion script for airshipctl for the specified shell (bash or zsh).


```
airshipctl completion SHELL [flags]
```

### Examples

```

# Save shell completion to a file
airshipctl completion bash > $HOME/.airship_completions

# Apply completions to the current shell
source <(airshipctl completion bash)

```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl](airshipctl.md)	 - A unified entrypoint to various airship components


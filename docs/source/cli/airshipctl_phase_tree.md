## airshipctl phase tree

Tree view of kustomize entrypoints of phase

### Synopsis

Summarized tree view of the kustomize entrypoints of a specific phase

```
airshipctl phase tree PHASE_NAME [flags]
```

### Examples

```

# yaml explorer of a phase with relative path
airshipctl phase tree /manifests/site/test-site/ephemeral/initinfra

#yaml explorer of a phase with phase name
airshipctl phase tree initinfra-ephemeral

```

### Options

```
  -h, --help   help for tree
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases


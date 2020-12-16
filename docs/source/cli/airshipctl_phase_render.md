## airshipctl phase render

Render phase documents from model

### Synopsis

Render phase documents from model

```
airshipctl phase render PHASE_NAME [flags]
```

### Examples

```

# Get all 'initinfra' phase documents containing labels "app=helm" and
# "service=tiller"
airshipctl phase render initinfra -l app=helm,service=tiller

# Get all documents containing labels "app=helm" and "service=tiller"
# and kind 'Deployment'
airshipctl phase render initinfra -l app=helm,service=tiller -k Deployment

```

### Options

```
  -a, --annotation string   filter documents by Annotations
  -g, --apiversion string   filter documents by API version
  -e, --executor            if set to true rendering will be performed by executor otherwise phase entrypoint will be rendered by kustomize, if entrypoint is not specified error will be returned
  -h, --help                help for render
  -k, --kind string         filter documents by Kinds
  -l, --label string        filter documents by Labels
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Manage phases


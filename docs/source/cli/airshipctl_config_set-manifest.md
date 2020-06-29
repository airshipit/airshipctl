## airshipctl config set-manifest

Manage manifests in airship config

### Synopsis

Create or modify a manifests in the airshipctl config file.


```
airshipctl config set-manifest NAME [flags]
```

### Examples

```

# Create a new manifest
airshipctl config set-manifest exampleManifest \
  --repo exampleRepo \
  --url https://github.com/site \
  --branch master \
  --primary \
  --sub-path exampleSubpath \
  --target-path exampleTargetpath

# Change the primary repo for manifest
airshipctl config set-manifest e2e \
  --repo exampleRepo \
  --primary

# Change the sub-path for manifest
airshipctl config set-manifest e2e \
  --sub-path treasuremap/manifests/e2e

# Change the target-path for manifest
airshipctl config set-manifest e2e \
  --target-path /tmp/e2e

```

### Options

```
      --branch string        the branch to be associated with repository in this manifest
      --commithash string    the commit hash to be associated with repository in this manifest
      --force                if set, enable force checkout in repository with this manifest
  -h, --help                 help for set-manifest
      --primary              if set, enable this repository as primary repository to be used with this manifest
      --repo string          the name of the repository to be associated with this manifest
      --sub-path string      the sub path to be set for this manifest
      --tag string           the tag to be associated with repository in this manifest
      --target-path string   the target path for to be set for this manifest
      --url string           the repository url to be associated with this manifest
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file


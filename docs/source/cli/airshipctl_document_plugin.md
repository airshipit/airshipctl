## airshipctl document plugin

Run as a kustomize exec plugin

### Synopsis

This command is meant to be used as a kustomize exec plugin.

The command reads the configuration file CONFIG passed as a first argument and
determines a particular plugin to execute. Additional arguments may be passed
to this command and can be used by the particular plugin.

CONFIG must be a structured kubernetes manifest (i.e. resource) and must have
'apiVersion' and 'kind' keys. If the appropriate plugin was not found, the
command returns an error.


```
airshipctl document plugin CONFIG [ARGS] [flags]
```

### Examples

```

# Perform a replacement on a deployment. Prior to running this command,
# the file '/tmp/replacement.yaml' should be created as follows:
---
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=nginx-latest].image

# The replacement can then be performed. Output defaults to stdout.
airshipctl document plugin /tmp/replacement.yaml

```

### Options

```
  -h, --help   help for plugin
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl document](airshipctl_document.md)	 - Manage deployment documents


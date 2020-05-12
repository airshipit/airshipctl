## airshipctl document render

Render documents from model

### Synopsis

Render documents from model

```
airshipctl document render [flags]
```

### Options

```
  -a, --annotation stringArray   filter documents by Annotations
  -g, --apiversion stringArray   filter documents by API version
  -f, --filter string            logical expression for document filtering
  -h, --help                     help for render
  -k, --kind stringArray         filter documents by Kinds
  -l, --label stringArray        filter documents by Labels
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl document](airshipctl_document.md)	 - Manage deployment documents


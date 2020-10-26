## airshipctl secret encrypt

Encrypt plain text yaml files representing Kubernetes objects consisting of sensitive configuration.

### Synopsis

Encrypt plain text yaml files representing Kubernetes objects consisting of sensitive configuration.

```
airshipctl secret encrypt [flags]
```

### Examples

```

# Encrypt all kubernetes objects in the manifests directory.
airshipctl secret encrypt

# Encrypt file from src and write to a different dst file
airshipctl secret encrypt \
	--src /tmp/manifests/target/secrets/qualified-secret.yaml \
	--dst /tmp/manifests/target/secrets/encrypted-qualified-secret.yaml

```

### Options

```
      --dst string          Path to the file or directory that has encrypted secrets for decryption. Defaults to src if empty.
  -h, --help                help for encrypt
      --kubeconfig string   Path to kubeconfig associated with cluster being managed
      --src string          Path to the file or directory that has secrets in plaintext that need to be encrypted. Defaults to the manifest location in airship config
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl secret](airshipctl_secret.md)	 - Manage secrets


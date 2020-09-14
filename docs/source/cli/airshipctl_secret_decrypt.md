## airshipctl secret decrypt

Decrypt encrypted yaml files into plaintext files representing Kubernetes objects consisting of sensitive data.

### Synopsis

Decrypt encrypted yaml files into plaintext files representing Kubernetes objects consisting of sensitive data.

```
airshipctl secret decrypt [flags]
```

### Examples

```

# Decrypt all encrypted files in the manifests directory.
airshipctl secret decrypt

# Decrypt encrypted file from src and write the plain text to a different dst file
airshipctl secret decrypt \
	--src /tmp/manifests/target/secrets/encrypted-qualified-secret.yaml \
	--dst /tmp/manifests/target/secrets/qualified-secret.yaml

```

### Options

```
      --dst string   Path to the file or directory to store decrypted secrets. Defaults to src if empty.
  -h, --help         help for decrypt
      --src string   Path to the file or directory that has secrets in encrypted text that need to be decrypted. Defaults to the manifest location in airship config
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl secret](airshipctl_secret.md)	 - Manage secrets


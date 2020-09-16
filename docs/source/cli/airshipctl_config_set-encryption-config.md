## airshipctl config set-encryption-config

Manage encryption configs in airship config

### Synopsis

Create or modify an encryption config in the airshipctl config file.

Encryption configs are local files or kubernetes secrets that are used to encrypt and decrypt kubernetes objects


```
airshipctl config set-encryption-config NAME [flags]
```

### Examples

```

# Create an encryption config with local gpg key source
airshipctl config set-encryption-config exampleConfig \
  --encryption-key path-to-encryption-key \
  --decryption-key path-to-encryption-key

# Create an encryption config with kube api server secret as the store to store encryption keys
airshipctl config set-encryption-config exampleConfig \
  --secret-name secretName \
  --secret-namespace secretNamespace

```

### Options

```
      --decryption-key-path string   the path to the decryption key file
      --encryption-key-path string   the path to the encryption key file
  -h, --help                         help for set-encryption-config
      --secret-name string           name of the secret consisting of the encryption and decryption keys
      --secret-namespace string      namespace of the secret consisting of the encryption and decryption keys
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file


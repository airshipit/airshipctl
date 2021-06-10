## airshipctl secret generate encryptionkey

Airshipctl command to generate a secure encryption key or passphrase

### Synopsis

Generates a secure encryption key or passphrase.

If regex arguments are passed the encryption key created would match the regular expression passed.


```
airshipctl secret generate encryptionkey [flags]
```

### Examples

```

Generates a secure encryption key or passphrase.
# airshipctl secret generate encryptionkey

Generates a secure encryption key or passphrase matching the regular expression
# airshipctl secret generate encryptionkey --regex Xy[a-c][0-9]!a*

```

### Options

```
  -h, --help           help for encryptionkey
      --limit int      limit number of characters for + or * regex (default 5)
      --regex string   regular expression string
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl secret generate](airshipctl_secret_generate.md)	 - Airshipctl command to generate secrets


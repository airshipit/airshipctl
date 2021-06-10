## airshipctl phase status

Airshipctl command to show status of the phase

### Synopsis


Get the status of a phase such as ephemeral-control-plane, target-initinfra etc...
To list the phases associated with a site, run 'airshipctl phase list'.


```
airshipctl phase status PHASE_NAME [flags]
```

### Examples

```

Status of initinfra phase
# airshipctl phase status ephemeral-control-plane

```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl phase](airshipctl_phase.md)	 - Airshipctl command to manage phases


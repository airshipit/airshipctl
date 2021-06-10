## airshipctl plan run

Airshipctl command to run plan

### Synopsis

Run a plan defined in the site manifest. Specify the plan using the mandatory parameter PLAN_NAME.
To get list of plans associated for a site, run 'airshipctl plan list'.


```
airshipctl plan run PLAN_NAME [flags]
```

### Examples

```

Run plan named iso
# airshipctl plan run iso

Perform a dry run of a plan
# airshipctl plan run iso --dry-run

```

### Options

```
      --dry-run                 simulate phase execution
  -h, --help                    help for run
      --wait-timeout duration   wait timeout
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl plan](airshipctl_plan.md)	 - Airshipctl command to manage plans


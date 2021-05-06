## airshipctl plan validate

Airshipctl command to validate plan

### Synopsis

Validate a plan defined in the site manifest. Specify the plan using the mandatory parameter PLAN_NAME.
To get list of plans associated for a site, run 'airshipctl plan list'.


```
airshipctl plan validate PLAN_NAME [flags]
```

### Examples

```

Validate plan named iso
# airshipctl plan validate iso

```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl plan](airshipctl_plan.md)	 - Airshipctl command to manage plans


## airshipctl cluster check-certificate-expiration

Check for expiring TLS certificates, secrets and kubeconfigs in the kubernetes cluster

### Synopsis

Displays a list of certificate expirations from both the management and
workload clusters, or in a self-managed cluster. Checks for TLS Secrets,
kubeconf secrets (which gets created while creating the workload cluster) and
also the node certificates present inside /etc/kubernetes/pki directory for
each node

```
airshipctl cluster check-certificate-expiration [flags]
```

### Examples

```

# To display all the expiring entities in the cluster
airshipctl cluster check-certificate-expiration --kubeconfig testconfig

# To display the entities whose expiration is within threshold of 30 days
airshipctl cluster check-certificate-expiration -t 30 --kubeconfig testconfig

# To output the contents to json (default operation)
airshipctl cluster check-certificate-expiration -o json --kubeconfig testconfig
or
airshipctl cluster check-certificate-expiration --kubeconfig testconfig

# To output the contents to yaml
airshipctl cluster check-certificate-expiration -o yaml --kubeconfig testconfig

# To output the contents whose expiration is within 30 days to yaml
airshipctl cluster check-certificate-expiration -t 30 -o yaml --kubeconfig testconfig

```

### Options

```
  -h, --help                help for check-certificate-expiration
      --kubeconfig string   Path to kubeconfig associated with cluster being managed
  -o, --output string       Convert output to yaml or json (default "json")
  -t, --threshold int       The max expiration threshold in days before a certificate is expiring. Displays all the certificates by default (default -1)
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl cluster](airshipctl_cluster.md)	 - Manage Kubernetes clusters


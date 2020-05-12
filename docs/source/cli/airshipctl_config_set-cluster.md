## airshipctl config set-cluster

Manage clusters

### Synopsis

Create or modify a cluster in the airshipctl config files.

Since a cluster can be either "ephemeral" or "target", you must specify
cluster-type when managing clusters.


```
airshipctl config set-cluster NAME [flags]
```

### Examples

```

# Set the server field on the ephemeral exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=ephemeral \
  --server=https://1.2.3.4

# Embed certificate authority data for the target exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=target \
  --client-certificate-authority=$HOME/.airship/ca/kubernetes.ca.crt \
  --embed-certs

# Disable certificate checking for the target exampleCluster
airshipctl config set-cluster exampleCluster
  --cluster-type=target \
  --insecure-skip-tls-verify

# Configure client certificate for the target exampleCluster
airshipctl config set-cluster exampleCluster \
  --cluster-type=target \
  --embed-certs \
  --client-certificate=$HOME/.airship/cert_file

```

### Options

```
      --certificate-authority string   path to a certificate authority
      --cluster-type string            the type of the cluster to add or modify
      --embed-certs                    if set, embed the client certificate/key into the cluster
  -h, --help                           help for set-cluster
      --insecure-skip-tls-verify       if set, disable certificate checking (default true)
      --server string                  server to use for the cluster
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
      --kubeconfig string    Path to kubeconfig associated with airshipctl configuration. (default "$HOME/.airship/kubeconfig")
```

### SEE ALSO

* [airshipctl config](airshipctl_config.md)	 - Manage the airshipctl config file


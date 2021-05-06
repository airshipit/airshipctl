## airshipctl cluster

Airshipctl command to manage kubernetes clusters

### Synopsis

Provides capabilities for interacting with a Kubernetes cluster,
such as getting status and deploying initial infrastructure.


### Options

```
  -h, --help   help for cluster
```

### Options inherited from parent commands

```
      --airshipconf string   Path to file for airshipctl configuration. (default "$HOME/.airship/config")
      --debug                enable verbose output
```

### SEE ALSO

* [airshipctl](airshipctl.md)	 - A unified entrypoint to various airship components
* [airshipctl cluster check-certificate-expiration](airshipctl_cluster_check-certificate-expiration.md)	 - Airshipctl command to check expiring TLS certificates, secrets and kubeconfigs in the kubernetes cluster
* [airshipctl cluster get-kubeconfig](airshipctl_cluster_get-kubeconfig.md)	 - Airshipctl command to retrieve kubeconfig for a desired cluster
* [airshipctl cluster list](airshipctl_cluster_list.md)	 - Airshipctl command to get and list defined clusters
* [airshipctl cluster rotate-sa-token](airshipctl_cluster_rotate-sa-token.md)	 - Airshipctl command to rotate tokens of Service Account(s)
* [airshipctl cluster status](airshipctl_cluster_status.md)	 - Retrieve statuses of deployed cluster components


## airshipctl cluster

Manage Kubernetes clusters

### Synopsis

This command provides capabilities for interacting with a Kubernetes cluster,
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
* [airshipctl cluster check-certificate-expiration](airshipctl_cluster_check-certificate-expiration.md)	 - Check for expiring TLS certificates, secrets and kubeconfigs in the kubernetes cluster
* [airshipctl cluster get-kubeconfig](airshipctl_cluster_get-kubeconfig.md)	 - Retrieve kubeconfig for a desired cluster
* [airshipctl cluster init](airshipctl_cluster_init.md)	 - Deploy cluster-api provider components
* [airshipctl cluster list](airshipctl_cluster_list.md)	 - Retrieve the list of defined clusters
* [airshipctl cluster move](airshipctl_cluster_move.md)	 - Move Cluster API objects, provider specific objects and all dependencies to the target cluster
* [airshipctl cluster rotate-sa-token](airshipctl_cluster_rotate-sa-token.md)	 - Rotate tokens of Service Accounts
* [airshipctl cluster status](airshipctl_cluster_status.md)	 - Retrieve statuses of deployed cluster components


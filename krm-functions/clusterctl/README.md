# Clusterctl

This is a KRM function which invokes
[clusterctl](https://github.com/kubernetes-sigs/cluster-api/tree/master/cmd/clusterctl)
with appropriate action and options.

## Function implementation

The function is implemented as an [image](image), and built using `make docker-image-clusterctl`.

### Function configuration

As input options, the KRM function receives a struct with command line options, configuration data and
repo components which is defined in airshipctl. See the `ClusterctlOptions` struct definition in v1alpha airshipctl API for the documentation.

## Function invocation

The function invoked by airshipctl command via `airshipctl phase run`:

    airshipctl phase run <phase_name>

if appropriate phase has Clusterctl executor defined.

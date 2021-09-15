# Applier

This is a KRM function which applies resources to k8s using
[cli-utils](https://github.com/kubernetes-sigs/cli-utils)
with appropriate options.

## Function implementation

The function is implemented as an [image](image), and built using `make docker-image-applier`.

### Function configuration

As input options, the KRM function receives a struct with apply options.
See the `ApplyConfig` struct definition in v1alpha1 airshipctl API for the documentation.

## Function invocation

The function invoked by airshipctl command via `airshipctl phase run`:

    airshipctl phase run <phase_name>

if appropriate phase has k8s_apply executor defined.

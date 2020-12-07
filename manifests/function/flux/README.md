# Flux

## How to Update

To update the version of upstream manifests used by a given function:

1. Update the versions (git refs) specified in the `dependencies` section
   of the Kptfile at the root of the function.
2. Run [`kpt pkg sync .`](https://github.com/GoogleContainerTools/kpt/blob/master/site/content/en/reference/pkg/sync/_index.md) from the root of the function.
3. Update any container image references in VariableCatalogues to match
   these new versions.
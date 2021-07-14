# Flux

## How to Update

##### Note: kpt version 1.0.0-beta.8 is required

To update the version of upstream manifests used by a given function:

1. Update the git refs specified in the `upstream` section of the Kptfile in each of the function's `upstream` directory's subdirectories, e.g. `base/upstream/policies/Kptfile`.
2. Save and commit the changes locally.
3. Run [`kpt pkg update`](https://kpt.dev/reference/cli/pkg/update/) from the directory containing the modified Kptfile.
4. After updating a package, all resulting changes must be committed before updating any additional package.
5. Update any container image references in VariableCatalogues to match these new versions.

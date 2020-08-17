Function: airshipctl-catalogues
===============================

This function defines some default VariableCatalogue resources,
which can be consumed and used (via ReplacementTransformer) to change the
versioning and resource locations used by functions in the airshipctl project.
More base catalogues will be added here in the future.

This catalogue can be used as-is to simply apply defaults, or a different
catalogue may be supplied (with the same ``versions-airshipctl`` name)
as a kustomize resource.  The catalogue in this function can also be
patched at the composite, type, or site level to reconfigure the versions.

The versions info falls under these keys:

* charts: Helm chart locations and versions

* files: image file (etc) locations and versions

* images: container image registries and versions

* kubernetes: a standalone key for the Kubernetes version to use

Versions that are defined for specific resources in specific functions
(e.g., container images) are categorized in the catalogue according
to the function and resource they will be applied to.
E.g., ``images.baremetal_operator.ironic.dnsmasq``.

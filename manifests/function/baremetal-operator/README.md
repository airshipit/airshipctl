Function: baremetal-operator
============================

This function defines a deployment of the Metal3 baremetal-operator,
including both the operator itself and Ironic.

Optional: a ``versions-airshipctl`` VariableCatalogue may be used to
override the default container images.
A base example for this catalogue can be found in the ``airshipctl-catalogues``
function.  If using the catalogue, apply the ``replacements/`` entrypoint
at the site level, as a Kustomize transformer.

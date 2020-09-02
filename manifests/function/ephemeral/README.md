Function: ephemeral
===================

This function defines the configuration for a bare metal ephemeral
bootstrapping image, which can be built via ``airshipctl image build``
and delivered over the WAN to a remote
host via redfish using ``airshipctl baremetal remotedirect``.

REQUIRED: a ``networking`` VariableCatalogue must be used to
override some Kubernetes networking configuration.
A base example for this catalogue can be found in the ``airshipctl-base-catalogues``
function.  If using the catalogue, apply the ``replacements/`` entrypoint
at the site level, as a Kustomize transformer.

Alternately, the entire text payload of the ephemeral secret may be overridden
via normal Kustomize patching.

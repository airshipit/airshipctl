Function: k8scontrol
====================

This function defines a base Clusterctl config that includes a collection
of available CAPI providers (under ``providers``) which are supported by
``airshipctl``.  It also provides a selection of those for a default Metal3
deployment (under ``init-options``).  The selected init-options may be
patched/overridden at the Type level, etc.

This function relies on CAPI variable substitution to supply versioned
container images to the CAPI components.  The Clusterctl objects
supplies defaults, and these can (optionally) be overridden either by
simple Kustomize patching, or by applying the ``replacements``
kustomization as a Kustomize transformer.  In the latter case,
an airshipctl versions catalogue must be supplied; please see the
``airshipctl-base-catalogues`` function for a base/example.

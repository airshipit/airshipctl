Function: k8scontrol
====================

This function defines a KubeADM and Metal3 control plane, including
Cluster, Metal3Cluster, KubeadmControlPlane, and Metal3MachineTemplate
resources.

Optional: a ``versions-airshipctl`` VariableCatalogue may be used to
override the default Kubernetes version and controlplane disk image.
A base example for this catalogue can be found in the ``airshipctl-base-catalogues``
function.  If using the catalogue, apply the ``replacements/`` entrypoint
at the site level, as a Kubernetes transformer.

Optional: a ``networking`` VariableCatalogue may be used to
override some Kubernetes networking configuration.
A base example for this catalogue can be found in the ``airshipctl-base-catalogues``
function.  If using the catalogue, apply the ``replacements/`` entrypoint
at the site level, as a Kustomize transformer.

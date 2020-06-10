Function: hostgenerator-m3
==========================

This function constructs a collection of Metal3 BareMetalHost resources,
along with associated configuration Secrets. It solves for a couple of things:

1. pulling the nitty gritty details for generating BMH into one reusable place,
2. allowing the site-specific details to be filled in via catalogues of values

This function leverages a couple of different plugins in sequence:
The airshipctl Replacement plugin, which pulls the site-specific data from
the catalogue documents into a Templater plugin configuration; and then
the airshipctl Templater plugin, which generates a variable number of
BMHs in a data-driven fashion.

To use this function, do the following:

* Supply a `common-networking-catalogue`, which outlines things that are
  typically common across hosts in a site, such as networking interfaces,
  DNS servers, and other networking info.
  Example: `manifests/type/gating/shared/catalogues/common-networking.yaml`

* Supply a `host-catalogue`, which contains host-specific data, such as
  IP addresses and BMC information.
  Example: `manifests/site/test-site/shared/catalogues/hosts.yaml`

* Supply a `host-generation-catalogue` for each `phase` that needs to
  deploy one or more BMHs.  This catalogue simply lists the specific
  hosts that should be deployed during that phase.
  Example: `manifests/site/test-site/ephemeral/bootstrap/hostgenerator/host-generation.yaml`

* If any per-host changes need to be made, they can be layered on top as
  site- or phase-specific Kustomize patches against the generated
  documents.  E.g, if one host has a different network interface name,
  or if different details need to be used during ISO bootstrapping
  and normal deployment.
  Example: `manifests/site/test-site/ephemeral/bootstrap/baremetalhost.yaml`

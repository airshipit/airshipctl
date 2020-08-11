Function: hardwareprofile-example
=================================

This function defines a hardware profile that can be consumed by the
hostgenerator-m3 function, and which has the same values as the default
profile defined in that function.  It serves as an example for how other
hardware profile functions can be created and consumed.

The `/replacements` kustomization contains a substution rule that injects
the profile into the hostgenerator BMH template.  Please see  the
`manifests/type/gating` type and `manifests/type/test-site`
kustomization.yamls to see how a hardwareprofile function can be wired in.


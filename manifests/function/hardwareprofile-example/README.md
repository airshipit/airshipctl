Function: hardwareprofile-example
=================================

This function defines a hardware profile that can be consumed by the
hostgenerator-m3 function. It serves as an example for how other
hardware profile functions can be created and consumed.

The `example` profile currently has fields for RAID and firmware configurations.
This is to provide as a reference for utilizing all the supported RAID levels
as well as all the supported firmware configurations.

For firmware configurations, the values from `example` profile are carried over
to the `default` profile of hostgenerator-m3. That is because same defaults
are exercised in metal3 baremetal-operator as well. See [bios-config spec]
However, for RAID configurations, since
there is no `default` profile, the template does *__not__* have any RAID fields.
Nevertheless, all the supported RAID configurations
have been listed in the `hardwareprofile.yaml` for your reference.

The `/replacements` kustomization contains a substitution rule that injects
the profile into the hostgenerator BMH template.  Please see  the
`manifests/type/gating` type and `manifests/site/test-site` site
kustomization.yamls to see how a hardwareprofile function can be wired in.

[bios-config spec]: https://github.com/metal3-io/metal3-docs/blob/master/design/baremetal-operator/bios-config.md

# Hardware Profile Guide

This document explains the functionality of RAID and firmware configuration
that is available in airshipctl. This will assist to author [Baremetal Host][3]
documents with added RAID and firmware functionality.

## The Host Generator M3 Function

Airshipctl has a `hostgenerator-m3` function which it uses to generate Metal3
[Baremetal Host][3] documents. This generator uses a `hosttemplate` to
templatize a BMH specification. It takes a set of parameters and a template as
inputs and provides BMH documents as outputs, making it easier to generate
a large set of BMH documents efficiently.

## The Example Hardware Profile

A Hardware Profile, in airshipctl terms, is a collection of parameters that
comprise a hardware level configuration of a server. Currently, it contains
RAID and firmware configurations. And later, this can be extended.

The `example` hardware profile is one such set, which is available as a
reference for all the supported parameters. You can modify this to your liking
to generate hardwareprofiles that suit your environment.

### Firmware Section

The firmware parameters supported, in the example profile
are as follows:

``` yaml
  firmware:
    sriovEnabled: false
    virtualizationDisabled: false
    simultaneousMultithreadingDisabled: false
```

These are the default values, you can adjust to your liking

### RAID Section

The RAID levels supported are 0, 1 and 1+0. Some examples
of using these levels in your configurations are given

``` yaml
  raid:
    hardwareRAIDVolumes:
    - name: "VirtualDisk1"
      level: "1+0"
      sizeGibibytes: 1024
      numberOfPhysicalDisks: 4
      rotational: False
    - name: "VirtualDisk2"
      level: "1"
      sizeGibibytes: 500
      numberOfPhysicalDisks: 2
      rotational: True
    - name: "VirtualDisk3"
      level: "0"
      sizeGibibytes: 500
      numberOfPhysicalDisks: 2
      rotational: True
    - name: "VirtualDisk4"
      level: "0"
      sizeGibibytes: 250
      numberOfPhysicalDisks: 1
      rotataional: False
```

For additional detail on these parameters, see the [Baremetal Host][1] API
documentation.

For more details on the example hardwareprofile, see [the repo][2].

[1]: https://github.com/metal3-io/baremetal-operator/blob/master/docs/api.md
[2]: https://opendev.org/airship/airshipctl/src/branch/master/manifests/function/hardwareprofile-example
[3]: https://github.com/metal3-io/baremetal-operator/tree/master/apis/metal3.io/v1alpha1

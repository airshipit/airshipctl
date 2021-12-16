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

The RAID levels supported are 0, 1, 5, 6, 1+0, 5+0, 6+0. Some examples
of using these levels in your configurations are given:

``` yaml
  raid:
    hardwareRAIDVolumes:
    - name: "VirtualDisk1"
      level: "0"
      sizeGibibytes: 2048
      numberOfPhysicalDisks: 2
      rotational: False
    - name: "VirtualDisk2"
      level: "1"
      controller: "RAID.Slot.5-1"
      physicalDisks:
      - "Disk.Bay.0:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.1:Enclosure.Internal.0-1:RAID.Slot.5-1"
    - name: "VirtualDisk3"
      level: "5"
      sizeGibibytes: 3000
      numberOfPhysicalDisks: 3
      rotational: True
    - name: "VirtualDisk4"
      level: "6"
      sizeGibibytes: 4000
      controller: "RAID.Slot.5-1"
      physicalDisks:
      - "Disk.Bay.0:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.1:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.2:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.3:Enclosure.Internal.0-1:RAID.Slot.5-1"
    - name: "VirtualDisk5"
      level: "1+0"
      sizeGibibytes: 4000
      numberOfPhysicalDisks: 4
    - name: "VirtualDisk6"
      level: "5+0"
      controller: "RAID.Slot.5-1"
      physicalDisks:
      - "Disk.Bay.0:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.1:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.2:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.3:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.4:Enclosure.Internal.0-1:RAID.Slot.5-1"
      - "Disk.Bay.5:Enclosure.Internal.0-1:RAID.Slot.5-1"
    - name: "VirtualDisk7"
      level: "6+0"
      numberOfPhysicalDisks: 8
      sizeGibibytes: 16000
      rotational: False
```
For additional detail on these parameters, see the [Baremetal Host][1] API
documentation.

Note that this has only been tested on Dell hardware.

For more details on the example hardwareprofile, see [the repo][2].

[1]: https://github.com/metal3-io/baremetal-operator/blob/master/docs/api.md
[2]: https://opendev.org/airship/airshipctl/src/branch/master/manifests/function/hardwareprofile-example
[3]: https://github.com/metal3-io/baremetal-operator/tree/master/apis/metal3.io/v1alpha1

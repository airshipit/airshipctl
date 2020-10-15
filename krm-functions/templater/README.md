# Templater function

This plugin is an implementation of a templater function written using `go` and uses the `kyaml`
and `airshipctl` libraries for parsing the input and writing the output.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

## Function invocation

The function is invoked by authoring a [Local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kustomize fn run local-resource/

The generated resources will appear in local-resource/

```
$ cat local-resource/*

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
...
```
### Configuration file format

`Templater` configuration resource is represented as a standard
k8s resource with Group, Version, Kind and Metadata header. Templater
configuration is defined using `template` and `values` fields with
following structure.

    values:
      hosts:
      - macAddress: 00:aa:bb:cc:dd
        name: node-1
      - macAddress: 00:aa:bb:cc:ee
        name: node-2
    template: |
      {{ range .hosts -}}
      ---
      apiVersion: metal3.io/v1alpha1
      kind: BareMetalHost
      metadata:
        name: {{ .name }}
      spec:
        bootMACAddress: {{ .macAddress }}
      {{ end -}}

`values` defines the substituion value as Map.
`template` defines the template with placeholders to substitue the value from the
           values Map

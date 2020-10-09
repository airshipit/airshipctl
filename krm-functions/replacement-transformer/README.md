# Replacement Transformer

This plugin is written in `go` and uses the `kyaml` and `airshipctl` libraries
for parsing the input and writing the output.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.
Function reads configuration, a collection of input resources, and performs values
replacement based on configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize config run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run `Replacement Transformer` with:

    kustomize fn run local-resource --dry-run

Value of `spec.version` in resource `KubeadmControlPlane` (`v1.18.6`) will be replaced
with value of `kubernetes` field defined in `VariableCatalogue` resource

## Configuration file format

`Replacement Transformer` configuration resource is represented as a standard
k8s resource with Group, Version, Kind and Metadata header. Replacement
configuration is defined under `replacements` field which contains a list of
object with following structure.

    source:
      objref:
        group: airshipit.org
        version: v1alpha1
        kind: Clusterctl
        name: resource-name
        namespace: capm3
      value: "string value"
      fieldref: {.data.host}
    target:
      objref:
        group: airshipit.org
        version: v1alpha1
        kind: KubeConfig
        name: resource-name
        namespace: capi-system
      fieldrefs:
        - {.config.kind}

* `source` defines where a substitution is from. It can from two different
kinds of sources from a field of one resource or from a string.
  * `objref` refers to a kubernetes object by Group, Version, Kind, Name and
  Namespace. Each field can be omitted or be an empty string.
  * `value` static string value to substitute into `target`.
  * `fieldref` JSON path to particular object field. This field essentially
  represents JSON query with syntax used in `kubectl` executed with
  flag `--jsonpath`. JSON path syntax end elements is defined by
  https://goessner.net/articles/JsonPath/
* `target` defines a substitution target.
  * `objref` specifies a set of resources. Any resource that matches
  intersection of all conditions (Group, Version, Kind, Name and Namespace) is
  included in this set.
  * `fieldrefs` list of JSON path strings which identify target field to
  substitute into. Field reference may have include pattern which is used as a
  replacement variable. For example in following query `{.metadata.name}%NAME%`
  string surrounded by `%` symbols (i.e. `NAME`) is considered as a pattern
  inside a field value identified by JSON path `metadata.name`. Therefore if
  value of `metadata.name` is `some-NAME-of-the-pod` only `NAME` substring is
  replaced with the string defined by substitution source.

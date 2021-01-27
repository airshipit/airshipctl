Function: generate-secrets-example
=================================

This function provide an example on how to generate secrets using templator
and variable catalogue. The generated secrets are usually of
`kind: VariableCatalogue`. These generated secrets then be used in
conjuction with `kind: ReplacementTransformer` to subsitute accordingly
in the site manifests. If the generated secrets needs to be deployed
on the cluster then define the secret as `kind: Secret` and appropriately
mark it with `deploy-k8s: true` annotation.

## Generating & Encrypting Secrets

Make a copy of this folder to the appropraite site for which secrets has to
be generated and then edit the [secret-generation.yaml](secret-generation.yaml)
with the required secret generation details.
For example refer to [generator](../../site/test-site/target/generator/) folder.

Once the secret definitions are in place in the site manifests, we can
add a new phase to generate secrets pointing to the folder in site manifests.
Below is an example of how to add phase to the [phases.yaml](../../phases/phases.yaml).

```
apiVersion: airshipit.org/v1alpha1
kind: Phase
metadata:
  name: secret-generate
config:
  executorRef:
    apiVersion: airshipit.org/v1alpha1
    kind: GenericContainer
    name: encrypter
  documentEntryPoint: target/generator
```

The executorRef is of `kind: GenericContainer` and should also have the
following definition in [executor.yaml](../../phases/executor.yaml)

```
---
apiVersion: airshipit.org/v1alpha1
kind: GenericContainer
metadata:
  name: encrypter
  labels:
    airshipit.org/deploy-k8s: "false"
kustomizeSinkOutputDir: "target/generator/results/generated"
spec:
  container:
    image: quay.io/aodinokov/sops:v0.0.3
    envs:
    - SOPS_IMPORT_PGP
    - SOPS_PGP_FP
config: |
  apiVersion: v1
  kind: ConfigMap
  data:
    cmd: encrypt
    unencrypted-regex: '^(kind|apiVersion|group|metadata)$'
```

The container spec in the `kind: GenericContainer` is specified with
sops spec so that the generated secrets would be encrypted and
then stored in the `kustomizeSinkOutputDir` directory. Sops uses pgp keys
and sops fingerprint key environment variable from the terminal to
perform encryption on the generated secrets.

## Steps to execute using airshipctl command

1. Sops environment variable has to be exported which will be
used for encryption. Download the sops key file. If you want to use
custom sops key copy it to the current location with filename as `key.asc`.

`curl -fsSL -o key.asc https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc`

2. Export key file and set corresponding fingerprint which will be
used for encryption.

`export SOPS_IMPORT_PGP="$(cat key.asc)" && export SOPS_PGP_FP="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4"`

3. Then run the airshipctl command

`airshipctl phase run <secret-generate>`

Once the command executes successfully, we can see the generated and
encrypted secrets will be placed in `kustomizeSinkOutputDir`.

## Generate Secrets without encryption(Not recommended)

In case if no encryption is required for the secrets then use the below
`kind: GenericContainer` definition in the [executor.yaml](../../phases/executor.yaml)

```
---
apiVersion: airshipit.org/v1alpha1
kind: GenericContainer
metadata:
  name: encrypter
  labels:
    airshipit.org/deploy-k8s: "false"
kustomizeSinkOutputDir: "target/generator/results/generated"
spec:
  container:
    image: quay.io/airshipit/templater:latest
config: |
  foo: bar
```

## Decrypt to read the secrets

To decrypt the secrets for readability purposes run the kustomize build
command on the generated secrets folder with the [kustomization.yaml](../../site/test-site/target/generator/results/kustomization.yaml) and [decrypt-secrets.yaml](../../site/test-site/target/generator/results/decrypt-secrets.yaml)
files in place in the same folder.

Kustomize command to decrypt:

`KUSTOMIZE_PLUGIN_HOME=$(pwd)/manifests SOPS_IMPORT_PGP=$(cat key.asc) kustomize build \ --enable_alpha_plugins \
manifests/site/test-site/target/generator/results`
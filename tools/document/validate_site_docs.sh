#!/bin/bash

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

: ${PROJECT_ROOT:=${PWD}}
: ${SITE:="test-workload"}
: ${CONTEXT:="kind-airship"}
: ${KUBECONFIG:="${HOME}/.airship/kubeconfig"}

: ${KUBECTL:="/usr/local/bin/kubectl"}
: ${KUSTOMIZE_PLUGIN_HOME:="${HOME}/.airship/kustomize-plugins"}
TMP=$(mktemp -d)

# Use the local project airshipctl binary as the default if it exists,
# otherwise use the one on the PATH
if [ -f "bin/airshipctl" ]; then
    AIRSHIPCTL_DEFAULT="bin/airshipctl"
else
    AIRSHIPCTL_DEFAULT="$(which airshipctl)"
fi

: ${AIRSHIPCONFIG:="${TMP}/config"}
: ${AIRSHIPKUBECONFIG:="${TMP}/kubeconfig"}
: ${AIRSHIPCTL:="${AIRSHIPCTL_DEFAULT}"}
ACTL="${AIRSHIPCTL} --airshipconf ${AIRSHIPCONFIG} --kubeconfig ${AIRSHIPKUBECONFIG}"

export KUSTOMIZE_PLUGIN_HOME
export KUBECONFIG

# TODO: use `airshipctl config` to do this once all the needed knobs are exposed
# The non-default parts are to set the targetPath and subPath appropriately,
# and to craft up cluster/contexts to avoid the need for automatic kubectl reconciliation
function generate_airshipconf {
    cluster=$1

    cat <<EOL > ${AIRSHIPCONFIG}
apiVersion: airshipit.org/v1alpha1
bootstrapInfo:
  default:
    builder:
      networkConfigFileName: network-config
      outputMetadataFileName: output-metadata.yaml
      userDataFileName: user-data
    container:
      containerRuntime: docker
      image: quay.io/airshipit/isogen:latest-debian_stable
      volume: /srv/iso:/config
    remoteDirect:
      isoUrl: http://localhost:8099/debian-custom.iso
clusters:
  ${CONTEXT}_${cluster}:
    clusterType:
      ${cluster}:
        bootstrapInfo: default
        clusterKubeconf: ${CONTEXT}_${cluster}
        managementConfiguration: default
contexts:
  ${CONTEXT}_${cluster}:
    contextKubeconf: ${CONTEXT}_${cluster}
    manifest: ${CONTEXT}_${cluster}
currentContext: ${CONTEXT}_${cluster}
kind: Config
managementConfiguration:
  default:
    insecure: true
    systemActionRetries: 30
    systemRebootDelay: 30
    type: redfish
manifests:
  ${CONTEXT}_${cluster}:
    primaryRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: https://opendev.org/airship/treasuremap
    subPath: manifests/site/${SITE}
    targetPath: .
users:
  ${CONTEXT}_${cluster}: {}
EOL
}

# Loop over all cluster types and phases for the given site
for cluster in ephemeral target; do
    # Clear out any CRDs left from testing of a previous cluster
    ${KUBECTL} --context ${CONTEXT} --kubeconfig ${KUBECONFIG} delete crd --all > /dev/null

    if [[ -d "manifests/site/${SITE}/${cluster}" ]]; then
        # Since we'll be mucking with the kubeconfig - make a copy of it and muck with the copy
        cp ${KUBECONFIG} ${AIRSHIPKUBECONFIG}
        # This is a big hack to work around kubeconfig reconciliation
        # change the cluster name (as well as context and user) to avoid kubeconfig reconciliation
        sed -i "s/${CONTEXT}/${CONTEXT}_${cluster}/" ${AIRSHIPKUBECONFIG}
        generate_airshipconf ${cluster}

        ${ACTL} cluster init
        phases="bootstrap initinfra "
        ignore=$(for i in $phases; do echo "-I $i "; done)
        phases+=$(ls $ignore manifests/site/${SITE}/${cluster}| grep -v "\.yaml$")
        for phase in $phases; do
            echo -e "\n*** Rendering ${cluster}/${phase}"

            # step 1: actually apply all crds in the phase
            ${ACTL} phase render ${phase} -k CustomResourceDefinition > ${TMP}/crds.yaml
            if [ -s ${TMP}/crds.yaml ]; then
                ${KUBECTL} --context ${CONTEXT} --kubeconfig ${KUBECONFIG} apply -f ${TMP}/crds.yaml
            fi

            # step 2: dry-run the entire phase
            ${ACTL} phase apply --dry-run ${phase}
        done
    fi
done

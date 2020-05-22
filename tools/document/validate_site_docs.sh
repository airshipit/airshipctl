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

# The root of the manifest structure to be validated.
# This corresponds to the targetPath in an airshipctl config
: ${MANIFEST_ROOT:="${PWD}"}
# The location of sites whose manifests should be validated.
# This are relative to MANIFEST_ROOT above, and correspond to
# the base of the subPath in an airshipctl config
: ${SITE_ROOT:="manifests/site"}

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
    subPath: ${SITE_ROOT}/${SITE}
    targetPath: ${MANIFEST_ROOT}
users:
  ${CONTEXT}_${cluster}: {}
EOL
}

function cleanup() {
    ${KIND} delete cluster --name airship
    rm -rf ${TMP}
}
trap cleanup EXIT

# Loop over all cluster types and phases for the given site
for cluster in ephemeral target; do
    if [[ -d "${MANIFEST_ROOT}/${SITE_ROOT}/${SITE}/${cluster}" ]]; then
        echo -e "\n**** Rendering phases for cluster: ${cluster}"
        # Start a fresh, empty kind cluster for validating documents
        ./tools/document/start_kind.sh

        # Since we'll be mucking with the kubeconfig - make a copy of it and muck with the copy
        cp ${KUBECONFIG} ${AIRSHIPKUBECONFIG}
        # This is a big hack to work around kubeconfig reconciliation
        # change the cluster name (as well as context and user) to avoid kubeconfig reconciliation
        sed -i "s/${CONTEXT}/${CONTEXT}_${cluster}/" ${AIRSHIPKUBECONFIG}
        generate_airshipconf ${cluster}

        ${ACTL} cluster init

        # A sequential list of potential phases.  A fancier attempt at this has been
        # removed since it was choking in certain cases and got to be more trouble than was worth.
        # This should be removed once we have a phase map that is smarter.
        # In the meantime, as new phases are added, please add them here as well.
        phases="bootstrap initinfra controlplane baremetalhost workers workload tenant"

        for phase in $phases; do
            # Guard against bootstrap or initinfra being missing, which could be the case for some configs
            if [ -d "${MANIFEST_ROOT}/${SITE_ROOT}/${SITE}/${cluster}/${phase}" ]; then
                echo -e "\n*** Rendering ${cluster}/${phase}"

                # step 1: actually apply all crds in the phase
                # TODO: will need to loop through phases in order, eventually
                # e.g., load CRDs from initinfra first, so they're present when validating later phases
                ${ACTL} phase render ${phase} -k CustomResourceDefinition > ${TMP}/${phase}-crds.yaml
                if [ -s ${TMP}/${phase}-crds.yaml ]; then
                    ${KUBECTL} --context ${CONTEXT} --kubeconfig ${KUBECONFIG} apply -f ${TMP}/${phase}-crds.yaml
                fi

                # step 2: dry-run the entire phase
                ${ACTL} phase apply --dry-run ${phase}
            fi
        done

        ${KIND} delete cluster --name airship
    fi
done

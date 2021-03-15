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

set -xe

# The root of the manifest structure to be validated.
# This corresponds to the targetPath in an airshipctl config
: ${MANIFEST_ROOT:="$(basename "${PWD}")/manifests"}
# The location of sites whose manifests should be validated.
# This are relative to MANIFEST_ROOT above
: ${SITE_ROOT:="$(basename "${PWD}")/manifests/site"}
: ${MANIFEST_REPO_URL:="https://review.opendev.org/airship/airshipctl"}
: ${SITE:="test-workload"}
: ${CONTEXT:="kind-airship"}
: ${AIRSHIPKUBECONFIG:="${HOME}/.airship/kubeconfig"}
: ${AIRSHIPKUBECONFIG_BACKUP:="${AIRSHIPKUBECONFIG}-backup"}


: ${KUBECTL:="/usr/local/bin/kubectl"}
TMP=$(mktemp -d)

# Use the local project airshipctl binary as the default if it exists,
# otherwise use the one on the PATH
if [ -f "bin/airshipctl" ]; then
  AIRSHIPCTL_DEFAULT="bin/airshipctl"
else
  AIRSHIPCTL_DEFAULT="$(which airshipctl)"
fi

: ${AIRSHIPCONFIG:="${TMP}/config"}
: ${AIRSHIPCTL:="${AIRSHIPCTL_DEFAULT}"}
ACTL="${AIRSHIPCTL} --airshipconf ${AIRSHIPCONFIG}"

export KUBECONFIG="${AIRSHIPKUBECONFIG}"

# TODO: use `airshipctl config` to do this once all the needed knobs are exposed
# The non-default parts are to set the targetPath appropriately,
# and to craft up cluster/contexts to avoid the need for automatic kubectl reconciliation
function generate_airshipconf() {
  cluster=$1

  cat <<EOL >${AIRSHIPCONFIG}
apiVersion: airshipit.org/v1alpha1
contexts:
  ${CONTEXT}_${cluster}:
    manifest: ${CONTEXT}_${cluster}
    managementConfiguration: default
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
    phaseRepositoryName: primary
    repositories:
      primary:
        checkout:
          branch: master
          commitHash: ""
          force: false
          tag: ""
        url: ${MANIFEST_REPO_URL}
    targetPath: ${MANIFEST_ROOT}
    metadataPath: manifests/site/${SITE}/metadata.yaml
EOL
}

function cleanup() {
  ${KIND} delete cluster --name $CLUSTER
  rm -rf ${TMP}

  if [ -f "${AIRSHIPKUBECONFIG_BACKUP}" ]; then
    echo "Restoring a backup copy of kubeconfig"
    cp "${AIRSHIPKUBECONFIG_BACKUP}" "${AIRSHIPKUBECONFIG}"
  fi
}
trap cleanup EXIT

if [ -f "${AIRSHIPKUBECONFIG}" ]; then
  echo "Making a backup copy of kubeconfig"
  cp "${AIRSHIPKUBECONFIG}" "${AIRSHIPKUBECONFIG_BACKUP}"
fi

generate_airshipconf "default"

phase_plans=$(airshipctl --airshipconf ${AIRSHIPCONFIG} plan list | grep "PhasePlan" | awk -F '/' '{print $2}' | awk '{print $1}')
for plan in $phase_plans; do
  # Perform static validation first, add support of all plans later
  if [ "$plan" = "phasePlan" ]; then
    airshipctl --airshipconf ${AIRSHIPCONFIG} plan validate $plan
  fi

  cluster_list=$(airshipctl --airshipconf ${AIRSHIPCONFIG} cluster list)
  # Loop over all cluster types and phases for the given site
  for cluster in $cluster_list; do
    echo -e "\n**** Rendering phases for cluster: ${cluster}"

    export CLUSTER="${cluster}"

    # Start a fresh, empty kind cluster for validating documents
    ./tools/document/start_kind.sh

    generate_airshipconf ${cluster}

    # A sequential list of potential phases.  A fancier attempt at this has been
    # removed since it was choking in certain cases and got to be more trouble than was worth.
    # This should be removed once we have a phase map that is smarter.
    # In the meantime, as new phases are added, please add them here as well.
    phases=$(airshipctl --airshipconf ${AIRSHIPCONFIG} phase list --plan $plan -c $cluster | grep Phase | awk -F '/' '{print $2}' | awk '{print $1}' || true)

    for phase in $phases; do
      # Guard against bootstrap or initinfra being missing, which could be the case for some configs
      echo -e "\n*** Rendering ${cluster}/${phase}"

      # step 1: actually apply all crds in the phase
      # TODO: will need to loop through phases in order, eventually
      # e.g., load CRDs from initinfra first, so they're present when validating later phases
      ${AIRSHIPCTL} --airshipconf ${AIRSHIPCONFIG} phase render ${phase} -s executor -k CustomResourceDefinition >${TMP}/${phase}-crds.yaml

      if [ -s ${TMP}/${phase}-crds.yaml ]; then
        ${KUBECTL} --context ${CLUSTER} apply -f ${TMP}/${phase}-crds.yaml
      fi

      # step 2: dry-run the entire phase
      ${ACTL} phase run --dry-run ${phase}

    done
    # Delete cluster kubeconfig
    rm ${KUBECONFIG}
    ${KIND} delete cluster --name $CLUSTER
  done
done

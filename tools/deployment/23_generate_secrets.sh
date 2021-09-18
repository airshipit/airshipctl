#!/usr/bin/env bash

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

echo "Generating secrets using airshipctl"
FORCE_REGENERATE=all airshipctl phase run secret-update

echo "Generating ~/.airship/kubeconfig"
export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-"https://review.opendev.org/airship/airshipctl"}
export EXTERNAL_KUBECONFIG=${EXTERNAL_KUBECONFIG:-""}
export SITE=${SITE:-"test-site"}
export WORKDIR="${AIRSHIP_CONFIG_MANIFEST_DIRECTORY}/$(basename ${AIRSHIP_CONFIG_PHASE_REPO_URL})"

if [[ -z "$EXTERNAL_KUBECONFIG" ]]; then
   # we want to take config from bundle - remove kubeconfig file so
   # airshipctl could regenerated it from kustomize
   [ -f "~/.airship/kubeconfig" ] && rm ~/.airship/kubeconfig
   # we need to use tmp file, because airshipctl uses it and fails
   # if we write directly
   airshipctl cluster get-kubeconfig > ~/.airship/tmp-kubeconfig
   mv ~/.airship/tmp-kubeconfig ~/.airship/kubeconfig
fi

# Validate that we generated everything correctly
decrypted1=$(airshipctl phase run secret-show)
if [[ -z "${decrypted1}" ]]; then
        echo "Got empty decrypted value"
        exit 1
fi

#remove default key from env
unset SOPS_IMPORT_PGP

echo "Sanity check 1: Check that we can decrypt everything with U1 and U2 creds"
# set user1 key
cp ${WORKDIR}/manifests/.private-keys/my.key ${WORKDIR}/manifests/.private-keys/my.key.old
cp ${WORKDIR}/manifests/.private-keys/exampleU1.key ${WORKDIR}/manifests/.private-keys/my.key

#make sure that decrypted valus stay the same
decrypted2=$(airshipctl phase run secret-show)
if [ "${decrypted1}" != "${decrypted2}" ]; then
        echo "reencrypted decrypted value is different from the original"
        exit 1
fi
# set user2 key
cp ${WORKDIR}/manifests/.private-keys/exampleU2.key ${WORKDIR}/manifests/.private-keys/my.key

#make sure that decrypted valus stay the same
decrypted2=$(airshipctl phase run secret-show)
if [ "${decrypted1}" != "${decrypted2}" ]; then
        echo "reencrypted decrypted value is different from the original"
        exit 1
fi

echo "Sanity check 2: reencrypt ephemeral site using U2 user"
ONLY_CLUSTERS=ephemeral airshipctl phase run secret-update

#make sure that decrypted valus stay the same
decrypted2=$(airshipctl phase run secret-show)
if [ "${decrypted1}" != "${decrypted2}" ]; then
        echo "reencrypted decrypted value is different from the original"
        exit 1
fi

echo "Sanity check 3: Try to reecnrypt ephemeral by user 3, who can't decrypt target"
cp ${WORKDIR}/manifests/.private-keys/exampleU3.key ${WORKDIR}/manifests/.private-keys/my.key
TOLERATE_DECRYPTION_FAILURES=true ONLY_CLUSTERS=ephemeral airshipctl phase run secret-update

decrypted3=$(TOLERATE_DECRYPTION_FAILURES=true airshipctl phase run secret-show)
if [ "${decrypted1}" == "${decrypted3}" ]; then
        echo "reencrypted decrypted value should be different because it has to contain unencrypted data"
        exit 1
fi

mv ${WORKDIR}/manifests/.private-keys/my.key.old ${WORKDIR}/manifests/.private-keys/my.key

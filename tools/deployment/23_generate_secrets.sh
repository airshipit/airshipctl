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
export SOPS_PGP_FP=${SOPS_PGP_FP_ENCRYPT:-"${SOPS_PGP_FP}"}
airshipctl phase run secret-generate

echo "Generating ~/.airship/kubeconfig"
export AIRSHIP_CONFIG_MANIFEST_DIRECTORY=${AIRSHIP_CONFIG_MANIFEST_DIRECTORY:-"/tmp/airship"}
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-"https://review.opendev.org/airship/airshipctl"}
export EXTERNAL_KUBECONFIG=${EXTERNAL_KUBECONFIG:-""}
export SITE=${SITE:-"test-site"}

if [[ -z "$EXTERNAL_KUBECONFIG" ]]; then
   # we want to take config from bundle - remove kubeconfig file so
   # airshipctl could regenerated it from kustomize
   [ -f "~/.airship/kubeconfig" ] && rm ~/.airship/kubeconfig
   # we need to use tmp file, because airshipctl uses it and fails
   # if we write directly
   airshipctl cluster get-kubeconfig > ~/.airship/tmp-kubeconfig
   mv ~/.airship/tmp-kubeconfig ~/.airship/kubeconfig
fi

#backward compatibility with previous behavior
if [[ -z "${SOPS_PGP_FP_ENCRYPT}" ]]; then
	#skipping sanity checks
	exit 0
fi

echo "Sanity check for secret-reencrypt phase"
decrypted1=$(airshipctl phase run secret-show)
if [[ -z "${decrypted1}" ]]; then
	echo "Got empty decrypted value"
	exit 1
fi

#make sure that generated file has right FP
grep "${SOPS_PGP_FP}" "${AIRSHIP_CONFIG_MANIFEST_DIRECTORY}/$(basename ${AIRSHIP_CONFIG_PHASE_REPO_URL})/manifests/site/$SITE/target/encrypted/results/generated/secrets.yaml"

#set new FP and reencrypt
export SOPS_PGP_FP=${SOPS_PGP_FP_REENCRYPT}
airshipctl phase run secret-reencrypt
#make sure that generated file has right FP
grep "${SOPS_PGP_FP}" "${AIRSHIP_CONFIG_MANIFEST_DIRECTORY}/$(basename ${AIRSHIP_CONFIG_PHASE_REPO_URL})/manifests/site/$SITE/target/encrypted/results/generated/secrets.yaml"

#make sure that decrypted valus stay the same
decrypted2=$(airshipctl phase run secret-show)
if [ "${decrypted1}" != "${decrypted2}" ]; then
	echo "reencrypted decrypted value is different from the original"
	exit 1
fi

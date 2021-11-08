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

set -ex

# Builds all of the images under tools/airship-in-a-pod and then configures
# the AIAP pod to never pull down images so it is sure to use the built
# images instead. This also requires a few other images to be pulled.


AIRSHIPCTL_REF=${AIRSHIPCTL_REF:-"master"}
export AIRSHIPCTL_REF
# Images that are required by airship-in-a-pod but not built
PULL_LIST="docker:stable-dind nginx quay.io/metal3-io/sushy-tools quay.io/airshipit/libvirt:aiap-v1"


pushd tools/airship-in-a-pod/ || exit

make -e images artifact-setup base infra-builder runner libvirt

for IMAGE in $PULL_LIST; do
    docker pull "$IMAGE"
done

# Now that we have built/pulled the images, lets change the imagePullPolicy to
# Never to be 100% confident they are used
echo "- op: replace
  path: \"/spec/containers/0/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/1/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/2/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/3/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/4/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/5/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/6/imagePullPolicy\"
  value: Never

- op: replace
  path: \"/spec/containers/7/imagePullPolicy\"
  value: Never

" >> examples/airshipctl/replacements.yaml

# Also replace the patchset to the environment variables
# while being sure to escape the slashes from the ref
echo "- op: replace
  path: \"/spec/containers/4/env/6/value\"
  value: $AIRSHIPCTL_REF

" >> examples/airshipctl/replacements.yaml



popd || exit
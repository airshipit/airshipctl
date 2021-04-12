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

# This interactive script makes a copy of an Airship reference site and turns it
# into a standalone project.

set -e

# We can override this to "../treasuremap/manifests/site/test-site" till stl3 is merged
: ${REFERENCE_SITE:="../treasuremap/manifests/site/reference-airship-core"}
REFERENCE_SITE_SHORT=$(echo ${REFERENCE_SITE} | sed "s|.*/||")
: ${REFERENCE_TYPE:="airship-core"}

# The branch/tag foo below is only needed till we work some kinks out of `airshipctl document pull`
# Note: for airshipctl and treasuremap, specify a tag OR a branch, with "" for the other
: ${AIRSHIPCTL_REF:="v2.0.0"} # We can override this to "v2.0" till v2.0.0 is tagged
: ${AIRSHIPCTL_REF_TYPE:="tag"} # We can override this to "branch" till v2.0.0 is tagged
: ${TREASUREMAP_REF:="v2.0.0"} # We can override this to "v2.0" till v2.0.0 is tagged
: ${TREASUREMAP_REF_TYPE:="tag"} # We can override this to "branch" till v2.0.0 is tagged

# Args expected by `airshipctl config set-manifest`
TREASUREMAP_CONF_REF="--${TREASUREMAP_REF_TYPE} ${TREASUREMAP_REF}"
AIRSHIPCTL_CONF_REF="--${AIRSHIPCTL_REF_TYPE} ${AIRSHIPCTL_REF}"

echo "This script will initialize a new Airship site definition, based on the"
echo "treasuremap reference manifests, in a new project that lives side-by-side"
echo "with airshipctl and treasuremap.  Please run this script from the airshipctl directory."
echo

if [[ -z "$SITE" || -z "$PROJECT" ]]; then
  read -p "Choose a name for your project: " PROJECT
  read -p "Choose a name for your site: " SITE
  SITE_LOCATION="../${PROJECT}/manifests/site/${SITE}"
  read -p "Creating '${SITE_LOCATION}', do you want to continue? (Y/N): " OK
  if ! [[ $OK == [yY] ]]; then
      echo "Site initialization cancelled"
      exit 1
  fi
else
  SITE_LOCATION="../${PROJECT}/manifests/site/${SITE}"
fi

if [[ -e "${SITE_LOCATION}" ]]; then
    echo "A site definition ${SITE_LOCATION} already exists, aborting"
    exit 2
fi

set -x

# TODO: replace with `airshipctl document pull` once tag/branch based pulls work
if [[ ! -e ../treasuremap ]]; then
  git clone https://opendev.org/airship/treasuremap ../treasuremap
fi
pushd .; cd ../treasuremap; git checkout ${TREASUREMAP_REF}; popd

# Initialize a new site from the treasuremap reference
mkdir -p "${SITE_LOCATION}"
cp -r ${REFERENCE_SITE}/* "${SITE_LOCATION}"

# Update kustomize references
find "${SITE_LOCATION}" -type f -exec sed -i \
    "s|type/${REFERENCE_TYPE}|../../treasuremap/manifests/type/${REFERENCE_TYPE}|g" {} +
find "${SITE_LOCATION}" -type f -exec sed -i \
    "s|\.\./function|../../../treasuremap/manifests/function|g" {} +


# Update metadata.yaml paths with the new site name
sed -i "s|${REFERENCE_SITE_SHORT}|${SITE}|g" "${SITE_LOCATION}/metadata.yaml"

# Set up airshipctl config file
# Note: set-context doesn't have an option to set managementConfiguration.  A problem?
if [[ ! -e ~/.airship/config ]]; then
  airshipctl config init
fi

airshipctl config set-manifest "${SITE}" \
    --repo airshipctl \
    --url "https://opendev.org/airship/airshipctl" \
    ${AIRSHIPCTL_CONF_REF}
airshipctl config set-manifest "${SITE}" \
    --repo treasuremap \
    --url "https://opendev.org/airship/treasuremap" \
    ${TREASUREMAP_CONF_REF}
airshipctl config set-manifest "${SITE}" \
    --repo primary \
    --url "https://github.com/my-organization/${PROJECT}" \
    --branch master
airshipctl config set-manifest "${SITE}" \
    --metadata-path "manifests/site/${SITE}/metadata.yaml" \
    --target-path "$(realpath ..)"

airshipctl config set-context ephemeral-cluster --manifest "${SITE}"
airshipctl config set-context target-cluster --manifest "${SITE}"
airshipctl config use-context ephemeral-cluster

# TODO: use airshipctl for this once `airshipctl config` supports it
sed -i "s|^    managementConfiguration:.*|    managementConfiguration: default|g" ~/.airship/config

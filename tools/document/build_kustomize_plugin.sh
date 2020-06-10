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

# Run this from the airshipctl project root
set -e

# This script builds a version of kustomize that is able to acces
# the ReplacementTransformer plugin.
# It assumes a build airshipctl binary in bin/, or if not,
# somewhere on the path.
if [ -f "bin/airshipctl" ]; then
    AIRSHIPCTL="bin/airshipctl"
else
    AIRSHIPCTL="$(which airshipctl)"
fi

: ${KUSTOMIZE_PLUGIN_HOME:="$HOME/.airship/kustomize-plugins"}
: ${AIRSHIP_TAG:="dev"}

# purge any previous airship plugins
rm -rf ${KUSTOMIZE_PLUGIN_HOME}/airshipit.org

# copy our plugin to the PLUGIN_ROOT, and give a kustomzie-friendly wrapper
for PLUGIN in ReplacementTransformer Templater; do
  PLUGIN_PATH=${KUSTOMIZE_PLUGIN_HOME}/airshipit.org/v1alpha1/$(echo ${PLUGIN} | awk '{print tolower($0)}')
  mkdir -p ${PLUGIN_PATH}
  cat > ${PLUGIN_PATH}/${PLUGIN} <<EOF
#!/bin/bash
\$(dirname \$0)/airshipctl document plugin "\$@"
EOF
  chmod +x ${PLUGIN_PATH}/${PLUGIN}
  cp -p ${AIRSHIPCTL} ${PLUGIN_PATH}/
done

# make a fake "variablecatalogue" no-op plugin, so kustomize
# doesn't barf on leftover catalogues that were used to construct other transformer configs
PLUGIN_PATH=${KUSTOMIZE_PLUGIN_HOME}/airshipit.org/v1alpha1/variablecatalogue
mkdir -p ${PLUGIN_PATH}
cat > ${PLUGIN_PATH}/VariableCatalogue <<EOF
#!/bin/bash
# This is a no-op kustomize plugin
EOF
chmod +x ${PLUGIN_PATH}/VariableCatalogue

# tell the user how to use this
echo -e "The airshipctl kustomize plugin has been installed.\nRun kustomize with:"
echo -e "KUSTOMIZE_PLUGIN_HOME=$KUSTOMIZE_PLUGIN_HOME \$GOPATH/bin/kustomize build --enable_alpha_plugins ..."

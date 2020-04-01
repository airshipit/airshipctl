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

set -ex

# Find all files of given format and add license if missing
add_license() {
  ext=$1
  template=$2
  # skipping license for testdata and manifests folders
  FILES=$(find -L . -name "*.${ext}" | grep -v "testdata" | grep -v "manifests")

  for each in $FILES
  do
    if ! grep -q 'Apache License' $each
    then
      cat tools/${template} $each >$each.new
      mv $each.new $each
    fi
  done
}

add_license_to_bash() {
  template=$1
  FILES=$(find -L . -name "*.sh" )
  NUM_OF_LINES=$(< "tools/$template" wc -l)

  for each in $FILES
  do
    if ! grep -q 'Apache License' $each
    then
      if head -1 $each | grep '^#!' > /dev/null
      then
        head -n 1 $each >>$each.new
        head -n $NUM_OF_LINES tools/$template >>$each.new
        tail -n+2 $each >>$each.new
        mv $each.new $each
      fi
    fi
  done
}

add_license 'go' 'license_go.txt'
add_license 'yaml' 'license_yaml.txt'
add_license 'yml' 'license_yaml.txt'
add_license_to_bash 'license_bash.txt'

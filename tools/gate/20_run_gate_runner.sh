#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

set -xeo pipefail

source tools/export_sops

export AIRSHIPCTL_WS=${AIRSHIPCTL_WS:-$PWD}
export AIRSHIP_CONFIG_PHASE_REPO_URL=${AIRSHIP_CONFIG_PHASE_REPO_URL:-$PWD}

ZUUL_JOBS_PATH=zuul.d/jobs.yaml
GATE_RUNNER_YAML_PATH=playbooks/airshipctl-gate-runner.yaml
OUTPUT_DIR=""
STOP_SCRIPT=""
SKIP_LIST=""
MUTE=0

show_help() {
cat << EOF
Usage: $0 [options]
Run set of deployments scripts for airshipctl

-h,   --help                 Display help
-s,   --stop-at    NUMBER    Specify script number where to stop execution
-p    --pass       LIST      Comma separated list of script numbers to skip
-o,   --output-dir DIRNAME   The output of each script will be saved in the specified directory in a separate file
-m,   --mute                 Mute the output from scripts

EOF
}

# read the options
options=$(getopt -o hmo:p:s: --long help,mute,output-dir:,pass:,stop-at: -- "$@")
if [ $? != 0 ] ; then echo "Failed to parse options...exiting." >&2 ; exit 1 ; fi
eval set -- "$options"

while true; do
  case "$1" in
  -s | --stop-at)
      STOP_SCRIPT="$2"
      shift 2
      ;;
  -p | --pass)
      SKIP_LIST="$2"
      shift 2
      ;;
  -o | --output-dir)
      OUTPUT_DIR="$2"
      mkdir -p $OUTPUT_DIR
      shift 2
      ;;
  -m | --mute )
      MUTE=1
      shift
      ;;
  -h | --help )
      show_help
      exit 0
      ;;
  -- )
      shift
      break
      ;;
  esac
done

SCRIPT_LIST=$(cat $ZUUL_JOBS_PATH | yq '.[] | select(.job.name == "airship-airshipctl-gate-script-runner") | .job.vars.gate_scripts[]' -c -r)
if [[ ! $SCRIPT_LIST ]]; then
  SCRIPT_LIST=$(cat $GATE_RUNNER_YAML_PATH | yq '.[]| select (.name=="airshipctl_gate_runner")| .tasks[]| select (.name=="set_default_gate_scripts")| .set_fact.gate_scripts_default[]' -c -r)
fi

SKIP_LIST=$(echo ${SKIP_LIST//,/ })

for script in $SCRIPT_LIST; do
    SCRIPT_NAME=$(awk -F "/" "{ print \$NF }" <<<$script)
    if [[ $SCRIPT_NAME =~ ([0-9]+) ]]; then
      SCRIPT_NUM="${BASH_REMATCH[1]}"
    fi
    if [[ " ${SKIP_LIST[@]} " =~ " ${SCRIPT_NUM} " ]]; then
      if [[ $STOP_SCRIPT ]] && [[ $SCRIPT_NAME =~ "${STOP_SCRIPT}_"* ]]; then
        break
      fi
      continue
    fi

    echo -e "\033[0;32m[ *** Run script $script *** ] \033[0m "
    cmd="sudo --preserve-env=AIRSHIPCTL_WS,AIRSHIP_CONFIG_PHASE_REPO_URL,SOPS_IMPORT_PGP,SOPS_PGP_FP $script"
    if [[ $OUTPUT_DIR ]]; then
      $cmd > ${OUTPUT_DIR}/${SCRIPT_NAME}.out 2>&1
    elif [[ "$MUTE" -eq "1" ]]; then
      $cmd > /dev/null 2>&1
    else
      $cmd
    fi
    if [[ $STOP_SCRIPT ]] && [[ $SCRIPT_NAME =~ "${STOP_SCRIPT}_"* ]]; then
      break
    fi
done

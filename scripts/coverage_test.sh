#!/bin/bash

check_coverage() {
  COVER_FILE=$1
  MIN_COVERAGE=$2
  coverage_float=$(go tool cover -func="${COVER_FILE}" | awk "/^total:/ { print \$3 }")
  coverage_int=${coverage_float%.*}
  if (( "${coverage_int}" < "${MIN_COVERAGE}" )) ; then
    echo "Coverage is at ${coverage_float}, but the required coverage is ${MIN_COVERAGE}"
    exit 1
  fi
}

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <cover_file> <required_coverage>"
  exit 1
fi

check_coverage $1 $2

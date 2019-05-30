#!/bin/bash

check_coverage() {
  cover_file=$1
  min_coverage=$2
  sed -i -e "\,airshipctl/pkg/apis\|airshipctl/pkg/client,d" "${cover_file}"
  coverage_float=$(go tool cover -func="${cover_file}" | awk "/^total:/ { print \$3 }")
  coverage_int=${coverage_float%.*}
  if (( "${coverage_int}" < "${min_coverage}" )) ; then
    echo "Coverage is at ${coverage_float}, but the required coverage is ${min_coverage}"
    exit 1
  else
    echo "Overall coverage: ${coverage_float} of statements"
  fi
}

if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <cover_file> <required_coverage>"
  exit 1
fi

check_coverage "$1" "$2"

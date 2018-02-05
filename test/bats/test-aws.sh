#!/bin/bash

set -eu -o pipefail
shopt -s extglob

declare -r DIR=$(cd "$(dirname "$0")" && pwd)
cd "${DIR}"
THIS_SCRIPT="${DIR}"/$(basename "$0")

TEST_FILTER="${1:-*}"

PATH="${PATH}:./bin/bats/libexec"

./bin/bats/bin/bats aws/*@(${TEST_FILTER})*.sh

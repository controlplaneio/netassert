#!/bin/bash

set -euo pipefail

if [[ "${DEBUG:-}" != "" ]]; then
  set -x
fi

if [[ "${TEST_YAML:-}" != "" ]]; then
  echo "${TEST_YAML}" | base64 -d >/code/test/test.yaml
fi

if [[ ! -f /code/test/test.yaml ]]; then
  echo "No /code/test/test.yaml provided"
  exit 1
fi

if [[ "${DEBUG:-}" != "" ]]; then
  pwd
  id
  ls -lasp /root/ /root/.ssh/ || true
  echo "/code/test/test.yaml:"
  cat /code/test/test.yaml
fi

[[ -d ${HOME}/.parallel ]] || mkdir -p ${HOME}/.parallel || true
[[ -f ${HOME}/.parallel/will-cite ]] || touch ~/.parallel/will-cite

exec "${@}"

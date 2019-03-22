#!/bin/bash

set -euo pipefail
DEBUG="${DEBUG:-0}"

if [[ "${DEBUG:-}" == "1" ]]; then
  set -x
fi

if [[ -e /var/run/docker.sock ]]; then
  groupadd docker

  # get gid of docker socket file
  DOCKER_SOCK_GID=$(ls -ng /var/run/docker.sock | cut -f3 -d' ')

  # get group of docker inside container
  DOCKER_GID=$(getent group docker | cut -f3 -d: || true)

  # if they don't match, adjust
  if [[ ! -z "${DOCKER_SOCK_GID}" && "${DOCKER_SOCK_GID}" != "${DOCKER_GID}" ]]; then
    groupmod -g "${DOCKER_SOCK_GID}" docker
  fi

  if ! groups netassert | grep -q docker; then
    usermod -aG docker netassert
  fi
fi

if [[ "${TEST_YAML:-}" != "" ]]; then
  echo "${TEST_YAML}" | base64 -d >/code/test/test.yaml
fi

if [[ ! -f /code/test/test.yaml ]]; then
  echo "No /code/test/test.yaml provided"
  exit 1
fi

if [[ "${DEBUG:-}" == "1" ]]; then
  pwd
  id
  ls -lasp \
    /home/netassert/ \
    /home/netassert/.ssh/ || true
  echo "/code/test/test.yaml:"
  cat /code/test/test.yaml
fi


[[ -d ${HOME}/.parallel ]] || mkdir -p ${HOME}/.parallel || true
[[ -f ${HOME}/.parallel/will-cite ]] || touch ~/.parallel/will-cite

gosu netassert bash -c "$(cat << EOF
[[ -d \${HOME}/.parallel ]] || mkdir -p \${HOME}/.parallel || true
[[ -f \${HOME}/.parallel/will-cite ]] || touch ~/.parallel/will-cite
EOF
)"

if [[ -d /tmp/.ssh ]]; then
  cp -a /tmp/.ssh /home/netassert/
fi
if [[ -L /home/netassert/.ssh/config ]]; then
 rm -f /home/netassert/.ssh/config
fi
if [[ -d /tmp/ssh-config ]]; then
  cp -af /tmp/ssh-config /home/netassert/.ssh/config
fi

chown netassert -R /home/netassert

# TODO(AJM) run without root
# sudo setcap cap_net_raw,cap_net_admin,cap_net_bind_service+eip /usr/bin/nmap
# exec gosu netassert "${@}"

# TODO(ajm) remove these root hacks when running rootlessly
if [[ "${DEBUG:-}" == "1" ]]; then
  whoami
  pwd
  ls -lasp
fi

mkdir -p ~/.ssh
cp /home/netassert/.ssh ~/ -a || true
chown "$(whoami)" -R ${HOME}/.ssh

if [[ "${DEBUG:-}" == "1" ]]; then
  ls -lasp ~/.ssh || true
  ls -lasp ~/.ssh/ || true

  # this file must exist on the host, but not in the container
  cat ~/.ssh/config || true
  echo "${@}" || true
fi

exec "${@}"

#!/usr/bin/env bash

load '../bin/bats-support/load'
load '../bin/bats-assert/load'

SSH_USER="${SSH_USER:-centos}"
TEMP_HOSTS_FILE="/tmp/aws-hosts-$(echo "${BASH_SOURCE[0]}" | sha256sum | awk '{print $1}').tmp"

AWS_TYPE_TAG_MANAGEMENT_SUFFIXES="JENKINS ARTIFACTORY ELK"

DEFAULT_ALLOW_CURL_TARGET="https://eu-west-2.console.aws.amazon.com"
DEFAULT_DENY_CURL_TARGET="https://httpbin.org/ip"
DEFAULT_USER_AGENT="Mozilla/5.0"
DEFAULT_TAG_ENV="Environment"
DEFAULT_ENVIRONMENT="ops"

_populate_hosts_file() {
  local SEARCH_HOST_TAGS="${1:-${AWS_TYPE_TAG_MANAGEMENT_SUFFIXES}}"
  local TAG_NAME="${2:-}"
  local SEARCH_PREFIX="${3:-}"

  [[ -z "${SEARCH_HOST_TAGS}" || -z "${TAG_NAME}" ]] && {
    echo "SEARCH_HOST_TAGS and TAG_NAME required"
    return 1
  }

  printf "" | tee "${TEMP_HOSTS_FILE}"

  for HOST in ${SEARCH_HOST_TAGS}; do
    SEARCH="${SEARCH_PREFIX}$(echo "${HOST}" | tr '[:upper:]' '[:lower:]')"
    VAR=$(_aws \
      ec2 describe-instances \
      --filter Name="tag:${TAG_NAME}",Values="${SEARCH}" \
      Name="tag:${DEFAULT_TAG_ENV}",Values="${DEFAULT_ENVIRONMENT}" \
      Name="instance-state-name",Values="running" \
      | jq '.Reservations[].Instances[].PrivateIpAddress'
    )
    local HOST_ENV_VAR=$(echo "${HOST//-/_}" | tr '[a-z]' '[A-Z]' | tr '.' '_')
    echo "HOST_${HOST_ENV_VAR}=${VAR}" >>"${TEMP_HOSTS_FILE}"
  done

  # ensure four lines with numbers (IPs) exist
  local TARGET_TAG_COUNT=$(wc --words <<< "${SEARCH_HOST_TAGS}")
  local ACTUAL_TAG_COUNT=$(grep -E '[0-9]' "${TEMP_HOSTS_FILE}" --count)
  [ "${ACTUAL_TAG_COUNT}" -ge "${TARGET_TAG_COUNT}" ]
}

populate_hosts_file() {
  _populate_hosts_file "${AWS_TYPE_TAG_MANAGEMENT_SUFFIXES}" 'Type' 'management-'
}

populate_hosts_file_k8s() {
  local HOSTS=(
    'ingress-external.ops.int.control-plane.io'
  )
  _populate_hosts_file "${HOSTS[*]}" 'Name'
}

_aws() {
  local AWS="aws"
  if [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
    AWS="aws --profile=${AWS_PROFILE}"
  fi
  ${AWS} "${@}"
}

_ssh() {
  local IP="${1}"
  shift
  _ssh_user "${IP}" "${SSH_USER}" "${@}"
}

_ssh_user() {
  local IP="${1}"
  local USER="${2}"
  shift 2
  local COMMAND="${@:-hostname}"
  ssh "${USER}@${IP}" \
    -o StrictHostKeyChecking=no \
    -o ConnectTimeout=3 \
    "${COMMAND}"
}

_curl_80() {
  local IP="${1}"
  _curl_request "http://${IP}:80"
}

_curl_443() {
  local IP="${1}"
  _curl_request "https://${IP}:443"
}

_curl_request() {
  local URL="${1}"
  local STATUS_CODE=$(curl \
    -s \
    -o /dev/null \
    -w "%{http_code}" \
    -A "$DEFAULT_USER_AGENT" \
    --compressed \
    --connect-timeout 5 \
    --fail \
    "${URL}"
  )
  [[ "${STATUS_CODE}" -ge 100 ]]
}

_curl_outbound() {
  local IP="${1}"
  shift
  local CURL_ARGS="${@:-${DEFAULT_ALLOW_CURL_TARGET}}"
  _ssh \
    "${IP}" \
    "curl \
      -A '${DEFAULT_USER_AGENT}' \
      --compressed --connect-timeout 5 \
      --fail \
      --insecure \
      ${CURL_ARGS}"
}

_curl_outbound_user() {
  local IP="${1}"
  local USER="${2}"
  shift 2
  local CURL_ARGS="${@:-${DEFAULT_ALLOW_CURL_TARGET}}"
  _ssh_user \
    "${IP}" \
    "${USER}" \
    "curl \
      -A '${DEFAULT_USER_AGENT}' \
      --compressed --connect-timeout 5 \
      --fail \
      --insecure \
      ${CURL_ARGS}"
}

_curl_outbound_k8s_401() {
  local IP="${1}"
  local K8S_API="${2}"
  local STATUS_CODE=$(_curl_outbound \
    "${IP}" \
    "-o /dev/null \
    -I \
    -w '%{http_code}' \
    ${K8S_API}"
  )
  [[ ${STATUS_CODE} -eq 401 ]]
}

_get_s3_logs_size() {
    local S3_BUCKET="${1}"
    local DATE="${2:-today}"

    _aws \
        s3 \
        ls \
        "${S3_BUCKET}/$(date -d "${DATE}" +%Y/%m/%d)/" \
    | awk '{BYTES_SUM+=$3} END {print BYTES_SUM}'
}

_enhanced_logging() {
  local IP="${1}"
  local log_state=$(_aws \
    ec2 describe-instances \
    --filter Name="private-ip-address",Values="${IP}" \
    Name="instance-state-name",Values="running" \
    | jq -c '.Reservations[].Instances[].Monitoring'
  )

  [[ "${log_state}" == '{"State":"enabled"}' ]]
}

_get_ec2_api_ips() {
  local RECORDS=$(for X in {1..30}; do \
    dig ec2.eu-west-1.amazonaws.com \
     | awk '{print $5}' \
     | grep -v amazonaws; \
   done \
   | sort -u
   )

   echo "${RECORDS}" | tr ' ' '\n'
}

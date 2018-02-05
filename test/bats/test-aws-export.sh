#!/bin/bash

set -eu -o pipefail
shopt -s extglob

declare -r DIR=$(cd "$(dirname "$0")" && pwd)
cd "${DIR}"
THIS_SCRIPT="${DIR}"/$(basename "$0")

AWS_PROFILE="${AWS_PROFILE:-dsab_london}"

AWS_ENTITIES="network-acls
security-groups
route-tables
vpc-peering-connections"

DATE=$(date +%Y-%m-%d_%H-%M)

_aws() {
  local AWS="aws"
  if [[ -z "${AWS_SECRET_ACCESS_KEY:-}" ]]; then
    AWS="aws --profile=${AWS_PROFILE}"
  fi
  ${AWS} "${@}"
}

for ENTITY in ${AWS_ENTITIES}; do
    [[ -z "${ENTITY}" ]] && continue
    echo "${ENTITY}"

    _aws ec2 \
        --output table \
        describe-"${ENTITY}" | tee "${DATE}_aws-${ENTITY}.txt"
done

echo "Exported AWS network data to:"
ls -d -1 $PWD/*_aws-*.txt
exit 0

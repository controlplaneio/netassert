#!/usr/bin/env bash

load '_helper'

# These tests assume an .ssh/config or ssh-agent configuration allowing SSH access to these servers.
# Enter AWS_PROFILE information to query the API for the servers under test
AWS_PROFILE="${AWS_PROFILE:-cp-ops}"

HOST_JENKINS=""
HOST_ARTIFACTORY=""

# ---

@test "Populates ENV vars from AWS API" {
  run populate_hosts_file
  assert_success
}

# ---

@test "Jenkins SSH" {
  source "${TEMP_HOSTS_FILE}"
  run _ssh "${HOST_JENKINS}"
  assert_success
}

@test "Jenkins port 80" {
  source "${TEMP_HOSTS_FILE}"
  run _curl_80 "${HOST_JENKINS}"
  assert_success
}

@test "Jenkins port 443" {
  source "${TEMP_HOSTS_FILE}"
  run _curl_443 "${HOST_JENKINS}"
  assert_success
}

@test "Jenkins Enhanced Logging" {
  source "${TEMP_HOSTS_FILE}"
  run _enhanced_logging "${HOST_JENKINS}"
  assert_success
}

# ---

@test "Artifactory SSH" {
  source "${TEMP_HOSTS_FILE}"
  run _ssh "${HOST_ARTIFACTORY}"
  assert_success
}

@test "Artifactory port 80" {
  source "${TEMP_HOSTS_FILE}"
  run _curl_80 "${HOST_ARTIFACTORY}"
  assert_success
}

@test "Artifactory port 443" {
  source "${TEMP_HOSTS_FILE}"
  run _curl_443 "${HOST_ARTIFACTORY}"
  assert_success
}

@test "Artifactory Enhanced Logging" {
  source "${TEMP_HOSTS_FILE}"
  run _enhanced_logging "${HOST_ARTIFACTORY}"
  assert_success
}

#!/usr/bin/env bash

check_param() {
  local name=$1
  local value=$(eval echo '$'$name)
  if [ "$value" == 'replace-me' ]; then
    echo "environment variable $name must be set"
    exit 1
  fi
}

check_for_rogue_vm() {
  local ip=$1
  set +e
  nc -vz -w 1 $ip 22
  status=$?
  set -e
  if [ "${status}" == "0" ]; then
    echo "aborting due to vm existing at ${ip}"
    exit 1
  fi
}

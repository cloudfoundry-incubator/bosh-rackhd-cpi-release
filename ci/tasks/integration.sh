#!/bin/bash

set -e

check_param() {
  local name=$1
  local value=$(eval echo '$'$name)
  if [ "$value" == 'replace-me' ]; then
    echo "environment variable $name must be set"
    exit 1
  fi
}

check_param ON_RACK_API_URI

cd bosh-cpi-release/
source .envrc

go install github.com/onsi/ginkgo/ginkgo

cd src/github.com/onrack/onrack-cpi
ginkgo -r

echo "ingegration test complete."

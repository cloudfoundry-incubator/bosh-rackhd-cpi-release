#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param ON_RACK_API_URI

cd bosh-cpi-release/
source .envrc

go install github.com/onsi/ginkgo/ginkgo

cd src/github.com/onrack/onrack-cpi
ginkgo -r

echo "ingegration test complete."

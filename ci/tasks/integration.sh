#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param RACKHD_API_URI

cd bosh-cpi-release/
source .envrc

go install github.com/onsi/ginkgo/ginkgo

cd src/github.com/rackhd/rackhd-cpi
ginkgo -r

echo "ingegration test complete."

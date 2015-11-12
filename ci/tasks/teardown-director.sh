#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param BOSH_VSPHERE_DIRECTOR

bosh -n target ${BOSH_VSPHERE_DIRECTOR}
echo 'yes' | bosh --user admin --password admin delete deployment bat-director

echo "Remove Release"
echo 'yes' | bosh --user admin --password admin delete release bosh-rackhd-cpi

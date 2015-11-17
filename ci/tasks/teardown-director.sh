#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param BOSH_VSPHERE_DIRECTOR
check_param DIRECTOR_DEPLOYMENT_NAME
check_param CPI_RELEASE_NAME

bosh -n target ${BOSH_VSPHERE_DIRECTOR}
echo 'yes' | bosh --user admin --password admin delete deployment ${DIRECTOR_DEPLOYMENT_NAME}

echo "Remove Release"
echo 'yes' | bosh --user admin --password admin delete release ${CPI_RELEASE_NAME}

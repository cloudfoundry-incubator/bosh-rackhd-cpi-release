#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param BOSH_DIRECTOR_PUBLIC_IP
check_param BOSH_DIRECTOR_PRIVATE_IP
check_param AGENT_PUBLIC_KEY
check_param DIRECTOR_PRIVATE_KEY_DATA
check_param PRIMARY_NETWORK_CIDR
check_param PRIMARY_NETWORK_GATEWAY
check_param PRIMARY_NETWORK_RANGE
check_param PRIMARY_NETWORK_MANUAL_IP
check_param SECONDARY_STATIC_IP
check_param BAT_SPEC

base_dir=${PWD}

mkdir -p ${base_dir}/keys
director_private_key_path="${base_dir}/keys/director"
echo "${DIRECTOR_PRIVATE_KEY_DATA}" > ${director_private_key_path}
chmod 600 ${director_private_key_path}

# checked by BATs environment helper (bosh-acceptance-tests.git/lib/bat/env.rb)
export BAT_DIRECTOR=${BOSH_DIRECTOR_PUBLIC_IP}
export BAT_STEMCELL=${base_dir}/stemcell/stemcell.tgz
export BAT_DEPLOYMENT_SPEC="${PWD}/bats-config.yml"
export BAT_VCAP_PASSWORD='c1oudc0w'
export BAT_DNS_HOST=${BOSH_DIRECTOR_PUBLIC_IP}
export BAT_INFRASTRUCTURE='rackhd'
export BAT_NETWORKING='manual'
export BAT_VCAP_PRIVATE_KEY=${director_private_key_path}

cd bats
./write_gemfile
bundle install

echo "using bosh CLI version..."
bosh version

echo "targeting bosh director at ${BOSH_DIRECTOR_PUBLIC_IP}"
bosh -n target ${BOSH_DIRECTOR_PUBLIC_IP}

cat > ${BAT_DEPLOYMENT_SPEC} <<EOF
---
cpi: rackhd
manifest_template_path: ${base_dir}/bosh-cpi-release/ci/templates/rackhd.yml.erb
properties:
  key_name:  bats
  use_static_ip: true
  second_static_ip: ${SECONDARY_STATIC_IP}
  public_key: ${AGENT_PUBLIC_KEY}
  pool_size: 1
  instances: 1
  uuid: $(bosh status --uuid)
  stemcell:
    name: bosh-openstack-kvm-ubuntu-trusty-go_agent-raw
    version: latest
  networks:
  - name: default
    static_ip: ${PRIMARY_NETWORK_MANUAL_IP}
    type: manual
    cidr: ${PRIMARY_NETWORK_CIDR}
    reserved: [${BOSH_DIRECTOR_PRIVATE_IP}]
    static: [${PRIMARY_NETWORK_RANGE}]
    gateway: ${PRIMARY_NETWORK_GATEWAY}
EOF

# create dev release
pushd ${PWD}/spec/system/assets/bat-release
rm -rf dev_releases
bosh create release --force --name bat
cp dev_releases/bat/* dev_releases/
bosh --user admin --password admin upload release
popd

echo "running the tests"
bundle exec rspec ${BAT_SPEC}

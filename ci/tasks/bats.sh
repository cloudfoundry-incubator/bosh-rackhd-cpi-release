#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param BOSH_DIRECTOR_PUBLIC_IP
check_param BOSH_DIRECTOR_PRIVATE_IP
check_param PRIVATE_KEY
check_param PRIMARY_NETWORK_CIDR
check_param PRIMARY_NETWORK_GATEWAY
check_param PRIMARY_NETWORK_RANGE
check_param PRIMARY_NETWORK_MANUAL_IP

export BAT_STEMCELL=${PWD}/stemcell/stemcell.tgz

cd bats
working_dir=${PWD}

# checked by BATs environment helper (bosh-acceptance-tests.git/lib/bat/env.rb)
export BAT_VCAP_PRIVATE_KEY=${PRIVATE_KEY}
export BAT_DIRECTOR=${BOSH_DIRECTOR_PUBLIC_IP}
export BAT_VCAP_PASSWORD='c1oudc0w'
export BAT_DNS_HOST=${BOSH_DIRECTOR_PUBLIC_IP}
export BAT_INFRASTRUCTURE='rackhd'
export BAT_NETWORKING='manual'
export BAT_DEPLOYMENT_SPEC="${working_dir}/bats-config.yml"

echo "using bosh CLI version..."
bosh version

echo "targeting bosh director at ${bosh_director_public_ip}"
bosh -n target ${BOSH_DIRECTOR_PUBLIC_IP}

cat > $BAT_DEPLOYMENT_SPEC <<EOF
---
cpi: rackhd
properties:
  key_name:  bats
  pool_size: 1
  instances: 1
  uuid: $(bosh status --uuid)
  stemcell:
    name: bosh-stemcell-3072-openstack-kvm-ubuntu-trusty-go_agent-raw
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

gem install bundle

./write_gemfile

bundle install

echo "running the tests"
bundle exec rspec spec

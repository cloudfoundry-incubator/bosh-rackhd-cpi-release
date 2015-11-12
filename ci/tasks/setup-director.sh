#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh
# source ./utils.sh

check_param BOSH_VSPHERE_DIRECTOR
check_param BOSH_DIRECTOR_PUBLIC_IP
check_param BOSH_DIRECTOR_PRIVATE_IP
check_param RACKHD_API_URI

echo "Check to see if director exists at" $BOSH_DIRECTOR_PUBLIC_IP
# check_for_rogue_vm $BOSH_DIRECTOR_PUBLIC_IP
echo "Director" $BOSH_DIRECTOR_PUBLIC_IP "does not exist"

bosh -n target ${BOSH_VSPHERE_DIRECTOR}
director_uuid=$(bosh status | grep UUID | tr -s ' ' | cut -d' ' -f3)
echo "Director UUID = "$director_uuid

echo "Upload Stemcell"
echo "Create Bosh Release Tarball"

pushd bosh-cpi-release/
# pushd ../../
  mkdir -p blobs/golang
  pushd blobs/golang
    wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
  popd

  $(bosh create release --force --with-tarball > create_release_output)
  release_tarball_path=$(cat create_release_output | grep 'Release tarball' | cut -d ' ' -f4)
  echo $release_tarball_path
  bosh --user admin --password admin upload release $release_tarball_path
popd

cat > "./director-manifest.yml" <<EOF
---
name: bat-director
director_uuid: ${director_uuid}

disk_pools:
- name: disks
  disk_size: 15_000

resource_pools:
- name: vms
  network: vm-network
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: 3120
  cloud_properties:
    cpu: 4
    ram: 4_096
    disk: 15_000
  env:
    bosh:
      # c1oudc0w is a default password for vcap user
      password: "$6$4gDD3aV0rdqlrKC$2axHCxGKIObs6tAmMTqYCspcdvQXh3JJcvWOY2WGb4SrdXtnCyNaWlrf3WEqvYR2MYizEGp3kMmbpwBC6jsHt0"

compilation:
  workers: 1
  network: vm-network
  reuse_compilation_vms: true
  cloud_properties:
    cpu: 2
    ram: 3_840
    disk: 8_096

update:
  canaries: 1
  max_in_flight: 1
  serial: false
  canary_watch_time: 1000-60000
  update_watch_time: 1000-60000

# replace all addresses with your network's IP range
#
# e.g. X.X.0.2 -> 10.0.0.2
networks:
  - name: vm-network
    type: manual
    subnets:
      - range: 192.168.10.0/24
        gateway: 192.168.10.1
        dns: [192.168.10.1]
        reserved: [192.168.10.2 - 192.168.10.210]
        static:
          - ${BOSH_DIRECTOR_PUBLIC_IP} - ${BOSH_DIRECTOR_PUBLIC_IP}
        cloud_properties: {name: 'VM Network'}
  - name: onrack-network
    type: manual
    subnets:
      - range: 172.31.128.0/22
        gateway: 172.31.128.1
        dns: [172.31.128.1]
        #reserved: [172.31.128.1-172.31.128.255]
        static:
          - ${BOSH_DIRECTOR_PRIVATE_IP} - ${BOSH_DIRECTOR_PRIVATE_IP}
        cloud_properties: {name: 'OnRack Network'}

releases:
  - name: bosh
    version: latest
  - name: rackhd-cpi
    version: latest

jobs:
- name: bosh
  instances: 1

  templates:
  - {name: nats, release: bosh}
  - {name: redis, release: bosh}
  - {name: postgres, release: bosh}
  - {name: blobstore, release: bosh}
  - {name: director, release: bosh}
  - {name: health_monitor, release: bosh}
  - {name: rackhd-cpi, release: rackhd-cpi}

  resource_pool: vms
  persistent_disk_pool: disks

  networks:
  - {name: vm-network, static_ips: [${BOSH_DIRECTOR_PUBLIC_IP}], default: [dns, gateway]}
  - {name: onrack-network, static_ips: [${BOSH_DIRECTOR_PRIVATE_IP}]}


  properties:
    nats:
      address: 127.0.0.1
      user: nats
      password: nats-password

    redis:
      listen_address: 127.0.0.1
      address: 127.0.0.1
      password: redis-password

    postgres: &db
      listen_address: 127.0.0.1
      host: 127.0.0.1
      user: postgres
      password: postgres-password
      database: bosh
      adapter: postgres

    blobstore:
      address: ${BOSH_DIRECTOR_PUBLIC_IP}
      port: 25250
      use_ssl: false
      provider: dav
      director: {user: director, password: director-password}
      agent: {user: agent, password: agent-password}

    director:
      address: 127.0.0.1
      name: my-bosh
      db: *db
      cpi_job: rackhd-cpi
      user_management:
        provider: local
        local:
          users:
          - {name: admin, password: admin}
          - {name: hm, password: hm-password}

    hm:
      director_account: {user: hm, password: hm-password}
      resurrector_enabled: true

    rackhd-cpi:
      apiserver: "${RACKHD_API_URI}"
      agent:
        mbus: "nats://nats:nats-password@${BOSH_DIRECTOR_PRIVATE_IP}:4222"
        blobstore:
          provider: dav
          options:
            endpoint: http://${BOSH_DIRECTOR_PRIVATE_IP}:25250
            user: agent
            password: agent-password

    agent: {mbus: "nats://nats:nats-password@${BOSH_DIRECTOR_PRIVATE_IP}:4222"}

    ntp: &ntp [0.pool.ntp.org, 1.pool.ntp.org]

cloud_provider:
  template: {name: rackhd-cpi, release: rackhd-cpi}
  mbus: "https://mbus:mbus-password@${BOSH_DIRECTOR_PRIVATE_IP}:6868"

  properties:
    rackhd-cpi:
      apiserver: "${RACKHD_API_URI}"
      agent:
        mbus: "https://mbus:Pbc7ssdfh8w2@0.0.0.0:6868"
        blobstore:
          provider: local
          options: {blobstore_path: /var/vcap/micro_bosh/data/cache}

EOF

bosh --user admin --password admin deployment ./director-manifest.yml
echo 'yes' | bosh --user admin --password admin deploy

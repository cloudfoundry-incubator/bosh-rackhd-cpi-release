#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param BOSH_VSPHERE_DIRECTOR
check_param BOSH_DIRECTOR_PUBLIC_IP
check_param BOSH_DIRECTOR_PRIVATE_IP
check_param BOSH_DIRECTOR_PUBLIC_KEY
check_param RACKHD_API_URL
check_param RACKHD_NETWORK
check_param CPI_RELEASE_NAME
check_param DIRECTOR_DEPLOYMENT_NAME
check_param BOSH_DIRECTOR_VM_NETWORK_NAME
check_param BOSH_DIRECTOR_VM_NETWORK_RESERVED
check_param BOSH_DIRECTOR_VM_NETWORK_DNS
check_param BOSH_DIRECTOR_VM_NETWORK_GATEWAY
check_param BOSH_DIRECTOR_VM_NETWORK_RANGE
check_param PASSWORD

base_dir=${PWD}

gem install bosh_cli --no-ri --no-rdoc
bosh -n target ${BOSH_VSPHERE_DIRECTOR}
bosh --non-interactive --user admin --password admin upload release ${base_dir}/bosh-release/release.tgz
bosh --non-interactive --user admin --password admin upload release ${base_dir}/redis-release/release.tgz

if bosh -n --user admin --password admin deployments | grep -F ${DIRECTOR_DEPLOYMENT_NAME}
then
  bosh -n --user admin --password admin delete deployment ${DIRECTOR_DEPLOYMENT_NAME}
fi

if bosh -n --user admin --password admin releases | grep -F ${CPI_RELEASE_NAME}
then
  bosh -n --user admin --password admin delete release ${CPI_RELEASE_NAME}
fi

cd bosh-cpi-release/
cat > config/private.yml << EOF
---
blobstore:
  s3:
    bucket_name: bosh-rackhd-cpi-blobs
EOF

bosh create release --force --name "${CPI_RELEASE_NAME}"
bosh --user admin --password admin upload release

public_key=$(echo ${BOSH_DIRECTOR_PUBLIC_KEY} | base64)
cat > "./director-manifest.yml" <<EOF
---
name: ${DIRECTOR_DEPLOYMENT_NAME}
director_uuid: $(bosh status --uuid)

disk_pools:
- name: disks
  disk_size: 15_000

resource_pools:
- name: vms
  network: vm-network
  stemcell:
    name: bosh-vsphere-esxi-ubuntu-trusty-go_agent
    version: latest
  cloud_properties:
    cpu: 4
    ram: 4_096
    disk: 15_000
  env:
    bosh:
      # c1oudc0w is a default password for vcap user
      password: ${PASSWORD}

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
      - range: ${BOSH_DIRECTOR_VM_NETWORK_RANGE}
        gateway: ${BOSH_DIRECTOR_VM_NETWORK_GATEWAY}
        dns: [${BOSH_DIRECTOR_VM_NETWORK_DNS}]
        reserved: [${BOSH_DIRECTOR_VM_NETWORK_RESERVED}]
        static:
          - ${BOSH_DIRECTOR_PUBLIC_IP} - ${BOSH_DIRECTOR_PUBLIC_IP}
        cloud_properties: {name: ${BOSH_DIRECTOR_VM_NETWORK_NAME}}
  - name: rackhd-network
    type: manual
    subnets:
      - range: 172.31.128.0/22
        gateway: 172.31.128.1
        dns: [172.31.128.1]
        #reserved: [172.31.128.1-172.31.128.255]
        static:
          - ${BOSH_DIRECTOR_PRIVATE_IP} - ${BOSH_DIRECTOR_PRIVATE_IP}
        cloud_properties: {name: '${RACKHD_NETWORK}'}

releases:
  - name: bosh
    version: latest
  - name: ${CPI_RELEASE_NAME}
    version: latest

jobs:
- name: bosh
  instances: 1

  templates:
  - {name: nats, release: bosh}
  - {name: postgres, release: bosh}
  - {name: blobstore, release: bosh}
  - {name: director, release: bosh}
  - {name: health_monitor, release: bosh}
  - {name: rackhd-cpi, release: ${CPI_RELEASE_NAME}}

  resource_pool: vms
  persistent_disk_pool: disks

  networks:
  - {name: vm-network, static_ips: [${BOSH_DIRECTOR_PUBLIC_IP}], default: [dns, gateway]}
  - {name: rackhd-network, static_ips: [${BOSH_DIRECTOR_PRIVATE_IP}]}


  properties:
    nats:
      address: 127.0.0.1
      user: nats
      password: nats-password

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
      generate_vm_passwords: false
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
      api_url: "${RACKHD_API_URL}"
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
  template: {name: rackhd-cpi, release: ${CPI_RELEASE_NAME}}
  mbus: "https://mbus:mbus-password@${BOSH_DIRECTOR_PRIVATE_IP}:6868"

  properties:
    rackhd-cpi:
      api_url: "${RACKHD_API_URL}"
      agent:
        mbus: "https://mbus:Pbc7ssdfh8w2@0.0.0.0:6868"
        blobstore:
          provider: local
          options: {blobstore_path: /var/vcap/micro_bosh/data/cache}

EOF

bosh --user admin --password admin deployment ./director-manifest.yml
echo 'yes' | bosh --user admin --password admin deploy

# hack
apt-get install sshpass
echo ${BOSH_DIRECTOR_PUBLIC_KEY} > director_key.pub
touch director_key
sshpass -p "c1oudc0w" ssh-copy-id -o StrictHostKeyChecking=no -i director_key.pub vcap@${BOSH_DIRECTOR_PUBLIC_IP}

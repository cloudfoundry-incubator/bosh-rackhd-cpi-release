#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh

check_param RACK_HD_API_URI
check_param AGENT_PUBLIC_KEY
check_param STATIC_IP
check_param GATEWAY
check_param RACKHD_CPI_LOG_LEVEL

AGENT_PUBLIC_KEY=$(echo ${AGENT_PUBLIC_KEY} | tr -d '\n' | tr -d ' ')

pushd ${PWD}/stemcell/
tar -zxvf stemcell.tgz
stemcell_path=${PWD}/image
popd

pushd ${PWD}/bosh-cpi-release/
source .envrc
go build github.com/rackhd/rackhd-cpi/rackhd-cpi

# Prepare config file
echo "Prepare config file"
cat > config_file <<EOF
{
  "apiserver": "${RACK_HD_API_URI}",
  "agent": {
    "blobstore":{
      "provider": "local",
      "options": {
        "blobstore_path": "/var/vcap/micro_bosh/data/cache"
      }
    },
    "mbus":"https://mbus:Pbc7ssdfh8w2@0.0.0.0:6868"
  }
}
EOF
cat config_file
config_path=${PWD}/config_file

# Prepare bosh network configuration
echo -e "\nPrepare bosh network configuration. Bosh network is"
cat > bosh_networks <<EOF
{
  "default": {
    "cloud_properties": {},
    "default": [
      "dns",
      "gateway"
    ],
    "dns": null,
    "gateway": "${GATEWAY}",
    "ip": "${STATIC_IP}",
    "netmask": "255.255.252.0",
    "type": "manual"
  }
}
EOF
cat bosh_networks

# Prepare create stemcell request
echo -e "\nPrepare create stemcell request. Request is"
cat > create_stemcell_request <<EOF
{"method": "create_stemcell", "arguments": ["${stemcell_path}"]}
EOF
cat create_stemcell_request

# Run create stemcell method
echo -e "\nRun create stemcell method"
response=$(cat create_stemcell_request | ./rackhd-cpi -configPath=${config_path})
echo ${response}
stemcell_id=$(echo ${response} | jq .result)
if [ -z "${stemcell_id}" ] || [ ${stemcell_id} == null ]; then
  echo "can not retrieve stemcell id"
  exit 1
fi
echo "got stemcell id: ${stemcell_id}"

# Prepare create vm request
echo -e "\nPrepare create vm request"
agent_id=$(uuidgen)
cat > create_vm_request <<EOF
{"method": "create_vm", "arguments": ["${agent_id}", ${stemcell_id}, {"public_key": "${AGENT_PUBLIC_KEY}"}, $(cat bosh_networks)]}
EOF
cat create_vm_request

# Run create vm method
echo -e "\nRun create vm method"
vm_cid=$(cat create_vm_request | ./rackhd-cpi -configPath=${config_path} | jq .result)
if [ -z "${vm_cid}" ] || [ ${vm_cid} == null ]; then
  echo "can not retrieve vm cid"
  exit 1
fi
echo "got vm cid: ${vm_cid}"

# Prepare delete vm request
echo -e "\nPrepare delete vm request"
cat > delete_vm_request <<EOF
{"method": "delete_vm", "arguments": [${vm_cid}]}
EOF
cat delete_vm_request

# Run delete vm method
echo -e "\nRun delete vm method"
cat delete_vm_request | ./rackhd-cpi -configPath=${config_path} 2>&1

# Prepare delete stemcell request
echo -e "\nPrepare delete stemcell request"
cat > delete_stemcell_request <<EOF
{"method": "delete_stemcell", "arguments": [${stemcell_id}]}
EOF
cat delete_stemcell_request

# Run delete stemcell method
echo -e "\nRun delete stemcell method"
cat delete_stemcell_request | ./rackhd-cpi -configPath=${config_path}

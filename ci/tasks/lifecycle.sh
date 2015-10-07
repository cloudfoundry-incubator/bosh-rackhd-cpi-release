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
check_param AGENT_PUBLIC_KEY
check_param STATIC_IP
check_param GATEWAY
check_param ONRACK_CPI_LOG_LEVEL

AGENT_PUBLIC_KEY=$(echo ${AGENT_PUBLIC_KEY} | tr -d '\n' | tr -d ' ')

sudo apt-get update
sudo apt-get -y install jq uuid-runtime

pushd ${PWD}/stemcell/
tar -zxvf stemcell.tgz
stemcell_path=${PWD}/image
popd

pushd ${PWD}/bosh-cpi-release/
source .envrc
go build github.com/onrack/onrack-cpi/onrack-cpi

cat > config_file <<EOF
{
  "apiserver": "${ON_RACK_API_URI}",
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

cat > create_stemcell_request <<EOF
{"method": "create_stemcell", "arguments": ["${stemcell_path}"]}
EOF
cat create_stemcell_request

response=$(cat create_stemcell_request | ./onrack-cpi -configPath=${config_path})
echo ${response}
stemcell_id=$(echo ${response} | jq .result)
if [ -z "${stemcell_id}" ] || [ ${stemcell_id} == null ]; then
  echo "can not retrieve stemcell id"
  exit 1
fi
echo "got stemcell id: ${stemcell_id}"

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

agent_id=$(uuidgen)
cat > create_vm_request <<EOF
{"method": "create_vm", "arguments": ["${agent_id}", ${stemcell_id}, {"public_key": "${AGENT_PUBLIC_KEY}"}, $(cat bosh_networks)]}
EOF
cat create_vm_request

vm_cid=$(cat create_vm_request | ./onrack-cpi -configPath=${config_path} | jq .result)
if [ -z "${vm_cid}" ] || [ ${vm_cid} == null ]; then
  echo "can not retrieve vm cid"
  exit 1
fi
echo "got vm cid: ${vm_cid}"

cat > delete_vm_request <<EOF
{"method": "delete_vm", "arguments": [${vm_cid}]}
EOF
cat delete_vm_request
cat delete_vm_request | ./onrack-cpi -configPath=${config_path} 2>&1

cat > delete_stemcell_request <<EOF
{"method": "delete_stemcell", "arguments": [${stemcell_id}]}
EOF
cat delete_stemcell_request
cat delete_stemcell_request | ./onrack-cpi -configPath=${config_path}

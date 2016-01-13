#!/usr/bin/env bash

set -e -x

source bosh-cpi-release/ci/tasks/utils.sh

check_param RACKHD_API_URI
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
  "apiserver": "${RACKHD_API_URI}:8080",
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

# Prepare has_vm method
echo -e "\nRun has_vm method"
cat > has_vm <<EOF
{"method": "has_vm", "arguments": [${vm_cid}]}
EOF
cat has_vm

# Run has_vm method
result=$(cat has_vm | ./rackhd-cpi --configPath=${config_path} | jq .result)
if [ -z "${result}" ]; then
  echo "invalid result returned from has_vm"
  exit 1
elif [ ${result} != true ]; then
  echo "vm ${vm_cid} not found"
  exit 1
fi
echo "vm ${vm_cid} found"

# Prepare metadata
echo -e "\nPrepare metadata"
cat > metadata <<EOF
{
  "director": "director-784430",
  "deployment": "redis",
  "job": "redis",
  "index": "1"
}
EOF
cat metadata

# Prepare set vm metadata request
echo -e "\nPrepare set vm metadata request"
cat > set_vm_metadata_request <<EOF
{"method": "set_vm_metadata", "arguments": [${vm_cid}, $(cat metadata)]}
EOF
cat set_vm_metadata_request

# Run set_vm_metadata
echo -e "\nRun set vm metadata method"
cat set_vm_metadata_request | ./rackhd-cpi -configPath=${config_path} 2>&1

# Prepare create_disk
echo -e "\nPrepare create disk"
cat > create_disk <<EOF
[
  100,
  {},
  ${vm_cid}
]
EOF
cat create_disk

# Prepare create disk request
echo -e "\nPrepare create disk request"
cat > create_disk_request <<EOF
{"method": "create_disk", "arguments": $(cat create_disk)}
EOF
cat create_disk_request

# Run create_disk
disk_cid=$(cat create_disk_request | ./rackhd-cpi --configPath=${config_path} | jq .result)
echo $disk_cid
if [ -z "${disk_cid}" ] || [ "${disk_cid}" == "null" ]; then
  echo "invalid result returned from create_disk"
  exit 1
fi
echo "disk ${disk_cid} found"

# Prepare has_disk method
echo -e "\nRun has_disk method"
cat > has_disk <<EOF
{"method": "has_disk", "arguments": [${disk_cid}]}
EOF
cat has_disk

# Run has_disk method
result=$(cat has_disk | ./rackhd-cpi --configPath=${config_path} | jq .result)
if [ -z "${result}" ] || [ "${result}" == "null" ]; then
  echo "invalid result returned from has_disk"
  exit 1
elif [ ${result} != true ]; then
  echo "disk ${disk_cid} not found"
  exit 1
fi
echo "disk ${disk_cid} found"

# Prepare get disks request
echo -e "\nPrepare get_disks request"
cat > get_disks_request <<EOF
{"method": "get_disks", "arguments": [${vm_cid}]}
EOF
cat get_disks_request

# Run get disks
get_disks_result=$(cat get_disks_request | ./rackhd-cpi --configPath=${config_path} | jq .result)
echo $get_disks_result
if echo $get_disks_result | grep -F ${disk_cid} && ! echo $get_disks_result | grep -F ","; then
  echo "disk ${disk_cid} found"
else
  echo "invalid result returned from get_disks"
  exit 1
fi

# Prepare attach disk request
echo -e "\nPrepare attach disk request"
cat > attach_disk_request <<EOF
{"method": "attach_disk", "arguments": [${vm_cid}, ${disk_cid}]}
EOF
cat attach_disk_request

# Run attach_disk
echo -e "\nRun attach disk method"
cat attach_disk_request | ./rackhd-cpi --configPath=${config_path} 2>&1

# Prepare detach disk request
echo -e "\nPrepare detach disk request"
cat > detach_disk_request <<EOF
{"method": "detach_disk", "arguments": [${vm_cid}, ${disk_cid}]}
EOF
cat detach_disk_request

# Run detach_disk
echo -e "\nRun detach disk method"
cat detach_disk_request | ./rackhd-cpi --configPath=${config_path} 2>&1

# Prepare delete disk request
echo -e "\nPrepare delete disk request"
cat > delete_disk_request <<EOF
{"method": "delete_disk", "arguments": [${disk_cid}]}
EOF
cat delete_disk_request

# Run delete_disk
echo -e "\nRun delete disk method"
cat delete_disk_request | ./rackhd-cpi --configPath=${config_path} 2>&1

# Prepare delete vm request
echo -e "\nPrepare delete vm request"
cat > delete_vm_request <<EOF
{"method": "delete_vm", "arguments": [${vm_cid}]}
EOF
cat delete_vm_request

# Run delete vm method
echo -e "\nRun delete vm method"
cat delete_vm_request | ./rackhd-cpi -configPath=${config_path} 2>&1

# Run has_vm method
result=$(cat has_vm | ./rackhd-cpi --configPath=${config_path} | jq .result)
if [ -z "${result}" ]; then
  echo "invalid result returned from has_vm"
  exit 1
elif [ ${result} != false ]; then
  echo "vm ${vm_cid} found after deletion"
  exit 1
fi
echo "vm ${vm_cid} deleted"

# Prepare delete stemcell request
echo -e "\nPrepare delete stemcell request"
cat > delete_stemcell_request <<EOF
{"method": "delete_stemcell", "arguments": [${stemcell_id}]}
EOF
cat delete_stemcell_request

# Run delete stemcell method
echo -e "\nRun delete stemcell method"
cat delete_stemcell_request | ./rackhd-cpi -configPath=${config_path}

#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh
source bosh-cpi-release/ci/tasks/lifecycle-helpers.sh

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
printf "%s\n" "Prepare config file"
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
printf "%s\n" "Prepare bosh network configuration. Bosh network is"
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

# Create VM First

do_create_stemcell ${config_path} ${stemcell_path}
printf "%s\n" "Stemcell ${stemcell_id} created"

disk_cid=""
do_create_vm ${config_path} ${stemcell_id} ${AGENT_PUBLIC_KEY} ${disk_cid}
printf "%s\n" "VM ${vm_cid} created"

do_has_vm ${config_path} ${vm_cid}
printf "%s\n" "Has_vm returned result ${has_vm_result} after creation"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "Invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != true ]; then
  printf "%s\n" "VM ${vm_cid} not found--aborting"
  exit 1
fi

do_set_vm_metadata ${config_path} ${vm_cid}

do_create_disk ${config_path} ${vm_cid}
printf "%s\n" "Persistent disk ${disk_cid} created"

do_has_disk ${config_path} ${disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "%s\n" "Invalid result returned from has_disk"
  exit 1
elif [ ${has_disk_result} != true ]; then
  printf "%s\n" "Disk ${disk_cid} not found"
  exit 1
fi

do_get_disks ${config_path} ${vm_cid} ${disk_cid}
printf "%s\n" "Result ${get_disks_result} returned from get_disks"
if echo $get_disks_result | grep -F ${disk_cid} && ! echo $get_disks_result | grep -F ","; then
  printf "%s\n" "Disk ${disk_cid} found"
else
  printf "%s\n" "Invalid result returned from get_disks"
  exit 1
fi

do_attach_disk ${config_path} ${vm_cid} ${disk_cid}

do_detach_disk ${config_path} ${vm_cid} ${disk_cid}

do_delete_disk ${config_path} ${disk_cid}

do_delete_vm ${config_path} ${vm_cid}

do_has_vm ${config_path} ${vm_cid}
printf "%s\n" "has_vm returned result ${has_vm_result} after deletion"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != false ]; then
  printf "%s\n" "vm ${vm_cid} found--aborting"
  exit 1
fi

# Create Disk First

do_create_disk ${config_path} "\"\""
if [ "${disk_cid}" == "null" ]; then
  printf "%s\n" "create_disk failed to create a disk"
  exit 1
fi
printf "%s\n" "Persistent disk ${disk_cid} created"

do_has_disk ${config_path} ${disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "Invalid result returned from has_disk\n"
  exit 1
elif [ ${has_disk_result} != true ]; then
  printf "%s\n" "Disk ${disk_cid} not found"
  exit 1
fi

do_create_vm ${config_path} ${stemcell_id} ${AGENT_PUBLIC_KEY} ${disk_cid}
printf "%s\n" "VM ${vm_cid} created"

do_has_vm ${config_path} ${vm_cid}
printf "%s\n" "Has_vm returned result ${has_vm_result} after creation"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "Invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != true ]; then
  printf "%s\n" "VM ${vm_cid} not found--aborting"
  exit 1
fi

do_set_vm_metadata ${config_path} ${vm_cid}

do_get_disks ${config_path} ${vm_cid} ${disk_cid}
printf "%s\n" "Result ${get_disks_result} returned from get_disks"
if echo $get_disks_result | grep -F ${disk_cid} && ! echo $get_disks_result | grep -F ","; then
  printf "%s\n" "Disk ${disk_cid} found"
else
  printf "%s\n" "Invalid result returned from get_disks"
  exit 1
fi

do_attach_disk ${config_path} ${vm_cid} ${disk_cid}

do_detach_disk ${config_path} ${vm_cid} ${disk_cid}

do_delete_disk ${config_path} ${disk_cid}

do_delete_vm ${config_path} ${vm_cid}

do_has_vm ${config_path} ${vm_cid}
printf "%s\n" "has_vm returned result ${has_vm_result} after deletion"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != false ]; then
  printf "%s\n" "vm ${vm_cid} found--aborting"
  exit 1
fi

do_delete_stemcell ${config_path} ${stemcell_id}

#!/usr/bin/env bash

set -e

source bosh-cpi-release/ci/tasks/utils.sh
source bosh-cpi-release/ci/tasks/lifecycle-helpers.sh

check_param CUSTOMIZED_STEMCELL_NAME
check_param RACKHD_API_URL
check_param AGENT_PUBLIC_KEY
check_param AGENT_STATIC_IP1
check_param AGENT_STATIC_IP2
check_param AGENT_STATIC_IP3
check_param GATEWAY
check_param RACKHD_CPI_LOG_LEVEL

AGENT_PUBLIC_KEY=$(echo ${AGENT_PUBLIC_KEY} | tr -d '\n' | tr -d ' ')

pushd ${PWD}/stemcell/
tar -zxvf ${CUSTOMIZED_STEMCELL_NAME}
stemcell_path=${PWD}/image
popd

pushd ${PWD}/bosh-cpi-release/
source .envrc
go build github.com/rackhd/rackhd-cpi/rackhd-cpi

# Prepare config file
printf "%s\n" "Prepare config file"
cat > config_file <<EOF
{
  "api_url": "${RACKHD_API_URL}",
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

do_create_stemcell ${config_path} ${stemcell_path}
printf "%s\n" "Stemcell ${stemcell_id} created"

# Create two VMs and one persistent disk on three separate nodes
blank_disk_cid="\"\""
blank_vm_cid="\"\""
node1_vm_cid_file_path="${PWD}/node1_vm_cid"
do_create_vm ${config_path} ${stemcell_id} ${AGENT_PUBLIC_KEY} ${blank_disk_cid} ${node1_vm_cid_file_path} ${AGENT_STATIC_IP1} &

node2_vm_cid_file_path="${PWD}/node2_vm_cid"
do_create_vm ${config_path} ${stemcell_id} ${AGENT_PUBLIC_KEY} ${blank_disk_cid} ${node2_vm_cid_file_path} ${AGENT_STATIC_IP2} &

node3_disk_cid_file_path="${PWD}/node3_disk_cid"
do_create_disk ${config_path} ${blank_vm_cid} ${node3_disk_cid_file_path} &

printf "%s\n" "Waiting for first round of creation processes to finish..."
wait
printf "%s\n" "Done waiting"
node1_vm_cid=$(cat ${node1_vm_cid_file_path})
node2_vm_cid=$(cat ${node2_vm_cid_file_path})
node3_disk_cid=$(cat ${node3_disk_cid_file_path})
printf "%s\n" "VM ${node1_vm_cid} created"
printf "%s\n" "VM ${node2_vm_cid} created"
printf "%s\n" "Persistent disk ${node3_disk_cid} created"

# Check states
do_has_vm ${config_path} ${node1_vm_cid}
printf "%s\n" "Has_vm returned result ${has_vm_result} after creation for VM ${node1_vm_cid}"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "Invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != true ]; then
  printf "%s\n" "VM ${node1_vm_cid} not found--aborting"
  exit 1
fi

do_has_vm ${config_path} ${node2_vm_cid}
printf "%s\n" "Has_vm returned result ${has_vm_result} after creation for VM ${node2_vm_cid}"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "Invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != true ]; then
  printf "%s\n" "VM ${node2_vm_cid} not found--aborting"
  exit 1
fi

do_has_disk ${config_path} ${node3_disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "%s\n" "Invalid result returned from has_disk"
  exit 1
elif [ ${has_disk_result} != true ]; then
  printf "%s\n" "Disk ${node3_disk_cid} not found"
  exit 1
fi

# Create persistent disk on node with VM
node1_disk_cid_file_path="${PWD}/node1_disk_cid"
do_create_disk ${config_path} ${node1_vm_cid} ${node1_disk_cid_file_path} &

# Create VM on node with persistent disk
node3_vm_cid_file_path="${PWD}/node3_vm_cid"
do_create_vm ${config_path} ${stemcell_id} ${AGENT_PUBLIC_KEY} ${node3_disk_cid} ${node3_vm_cid_file_path} ${AGENT_STATIC_IP3} &

# do_set_vm_metadata
do_set_vm_metadata ${config_path} ${node2_vm_cid}

printf "%s\n" "Waiting for second round of creation processes to finish..."
wait
printf "%s\n" "Done waiting"
node1_disk_cid=$(cat ${node1_disk_cid_file_path})
node3_vm_cid=$(cat ${node3_vm_cid_file_path})
printf "%s\n" "VM ${node3_vm_cid} created"
printf "%s\n" "Persistent disk ${node1_disk_cid} created"

do_get_disks ${config_path} ${node1_vm_cid}
printf "%s\n" "Result ${get_disks_result} returned from get_disks"
if echo $get_disks_result | grep -F ${node1_disk_cid} && ! echo $get_disks_result | grep -F ","; then
  printf "%s\n" "Disk ${node1_disk_cid} found"
else
  printf "%s\n" "Invalid result returned from get_disks"
  exit 1
fi

do_has_vm ${config_path} ${node3_vm_cid}
printf "%s\n" "Has_vm returned result ${has_vm_result} after creation for VM ${node2_vm_cid}"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "Invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != true ]; then
  printf "%s\n" "VM ${node3_vm_cid} not found--aborting"
  exit 1
fi

do_has_disk ${config_path} ${node1_disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "%s\n" "Invalid result returned from has_disk"
  exit 1
elif [ ${has_disk_result} != true ]; then
  printf "%s\n" "Disk ${node1_disk_cid} not found"
  exit 1
fi

do_attach_disk ${config_path} ${node1_vm_cid} ${node1_disk_cid}

do_attach_disk ${config_path} ${node3_vm_cid} ${node3_disk_cid}

do_detach_disk ${config_path} ${node1_vm_cid} ${node1_disk_cid}

# Delete all created VMs and Disks

printf "%s\n" "Deleting disk ${node1_disk_cid}"
do_delete_disk ${config_path} ${node1_disk_cid}

printf "%s\n" "Deleting VM ${node3_vm_cid}"
do_delete_vm ${config_path} ${node3_vm_cid} &

printf "%s\n" "Deleting VM ${node2_vm_cid}"
do_delete_vm ${config_path} ${node2_vm_cid} &

printf "%s\n" "Deleting VM ${node1_vm_cid}"
do_delete_vm ${config_path} ${node1_vm_cid} &

printf "%s\n" "Waiting for VM deletion to finish..."
wait
printf "%s\n" "Done waiting"

printf "%s\n" "Deleting disk ${node3_disk_cid}"
do_delete_disk ${config_path} ${node3_disk_cid}

do_has_vm ${config_path} ${node1_vm_cid}
printf "%s\n" "has_vm returned result ${has_vm_result} after deletion"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != false ]; then
  printf "%s\n" "vm ${node1_vm_cid} found--aborting"
  exit 1
fi

do_has_vm ${config_path} ${node2_vm_cid}
printf "%s\n" "has_vm returned result ${has_vm_result} after deletion"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != false ]; then
  printf "%s\n" "vm ${node2_vm_cid} found--aborting"
  exit 1
fi

do_has_vm ${config_path} ${node3_vm_cid}
printf "%s\n" "has_vm returned result ${has_vm_result} after deletion"
if [ -z "${has_vm_result}" ]; then
  printf "%s\n" "invalid result returned from has_vm"
  exit 1
elif [ ${has_vm_result} != false ]; then
  printf "%s\n" "vm ${node3_vm_cid} found--aborting"
  exit 1
fi

do_has_disk ${config_path} ${node1_disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "%s\n" "Invalid result returned from has_disk"
  exit 1
elif [ ${has_disk_result} != false ]; then
  printf "%s\n" "Disk ${node1_disk_cid} found--aborting"
  exit 1
fi

do_has_disk ${config_path} ${node3_disk_cid}
printf "%s\n" "Result ${has_disk_result} returned from has_disk"
if [ -z "${has_disk_result}" ] || [ "${has_disk_result}" == "null" ]; then
  printf "%s\n" "Invalid result returned from has_disk"
  exit 1
elif [ ${has_disk_result} != false ]; then
  printf "%s\n" "Disk ${node3_disk_cid} found--aborting"
  exit 1
fi

do_delete_stemcell ${config_path} ${stemcell_id}

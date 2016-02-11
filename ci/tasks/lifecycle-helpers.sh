#!/usr/bin/env bash
set -e

do_create_stemcell() {
  local config_path=$1
  local stemcell_path=$2

  # Prepare create stemcell request
  printf "%s\n" "Prepare create stemcell request. Request is"
  cat > create_stemcell_request <<EOF
{"method": "create_stemcell", "arguments": ["${stemcell_path}"]}
EOF
  cat create_stemcell_request

  # Run create stemcell method
  printf "%s\n" "Run create stemcell method"
  local response=$(cat create_stemcell_request | ./rackhd-cpi -configPath=${config_path})
  echo $response
  stemcell_id=$(echo ${response} | jq .result)
  if [ -z "${stemcell_id}" ] || [ ${stemcell_id} == null ]; then
    echo "can not retrieve stemcell id"
    exit 1
  fi
}

do_create_vm() {
  local config_path=$1
  local stemcell_id=$2
  local AGENT_PUBLIC_KEY=$3
  local disk_cid=$4
  local vm_cid_file_path=$5
  local agent_static_ip=$6

  # Prepare bosh network configuration
  printf "%s\n" "Prepare bosh network configuration. Bosh network is"
  local bosh_network_file="bosh_network_file_$(uuidgen)"
  cat > ${bosh_network_file} <<EOF
{
  "default": {
    "cloud_properties": {},
    "default": [
      "dns",
      "gateway"
    ],
    "dns": null,
    "gateway": "${GATEWAY}",
    "ip": "${agent_static_ip}",
    "netmask": "255.255.252.0",
    "type": "manual"
  }
}
EOF
  cat ${bosh_network_file}

  # Prepare create vm request
  printf "%s\n" "Prepare create vm request"
  local agent_id=$(uuidgen)
  local request_file="create_vm_request_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "create_vm",
  "arguments": [
    "${agent_id}",
    ${stemcell_id},
    {
      "public_key": "${AGENT_PUBLIC_KEY}"
    },
    $(cat ${bosh_network_file}),
    [${disk_cid}]
  ]
}
EOF
  cat ${request_file}

  # Run create vm method
  printf "%s\n" "Run create vm method"
  vm_cid=$(cat ${request_file} | ./rackhd-cpi -configPath=${config_path} | jq .result)
  if [ -z "${vm_cid}" ] || [ ${vm_cid} == null ]; then
    echo "can not retrieve vm cid"
    exit 1
  fi

  if [ ! -z ${vm_cid_file_path} ]; then
    echo ${vm_cid} > ${vm_cid_file_path}
  fi
}

do_has_vm() {
  local config_path=$1
  local vm_cid=$2

  # Prepare has_vm method
  printf "%s\n" "Run has_vm method"
  local request_file="has_vm_request_$(uuidgen)"

  cat > has_vm <<EOF
{"method": "has_vm", "arguments": [${vm_cid}]}
EOF
  cat has_vm

  # Run has_vm method
  has_vm_result=$(cat has_vm | ./rackhd-cpi --configPath=${config_path} | jq .result)
}

do_set_vm_metadata() {
  local config_path=$1
  local vm_cid=$2

  # Prepare metadata
  printf "%s\n" "Prepare metadata"
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
  printf "%s\n" "Prepare set vm metadata request"
  cat > set_vm_metadata_request <<EOF
{"method": "set_vm_metadata", "arguments": [${vm_cid}, $(cat metadata)]}
EOF
  cat set_vm_metadata_request

  # Run set_vm_metadata
  printf "%s\n" "Run set vm metadata method"
  cat set_vm_metadata_request | ./rackhd-cpi -configPath=${config_path} 2>&1
}

do_create_disk() {
  local config_path=$1
  local vm_cid=$2
  local disk_cid_file_path=$3

  # Prepare create_disk
  printf "%s\n" "Prepare create disk"
  local request_file="create_disk_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "create_disk",
  "arguments": [
    100,
    {},
    ${vm_cid}
  ]
}
EOF
  cat ${request_file}

  # Run create_disk
  disk_cid=$(cat ${request_file} | ./rackhd-cpi --configPath=${config_path} | jq .result)

  if [ ! -z ${disk_cid_file_path} ]; then
    echo ${disk_cid} > ${disk_cid_file_path}
  fi
}

do_has_disk() {
  local config_path=$1
  local disk_cid=$2

  # Prepare has_disk method
  printf "%s\n" "Prepare has_disk method"
  local request_file="has_disk_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "has_disk",
  "arguments": [
    ${disk_cid}
  ]
}
EOF
  cat ${request_file}

  # Run has_disk method
  has_disk_result=$(cat ${request_file} | ./rackhd-cpi --configPath=${config_path} | jq .result)
}

do_get_disks() {
  local config_path=$1
  local vm_cid=$2

  # Prepare get disks request
  printf "%s\n" "Prepare get_disks request"
  local request_file="get_disks_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "get_disks",
  "arguments": [
    ${vm_cid}
  ]
}
EOF
  cat ${request_file}

  # Run get disks
  get_disks_result=$(cat ${request_file} | ./rackhd-cpi --configPath=${config_path} | jq .result)
  echo $get_disks_result
}

do_attach_disk() {
  local config_path=$1
  local vm_cid=$2
  local disk_cid=$3

  # Prepare attach disk request
  printf "%s\n" "Prepare attach disk request"
  local request_file="attach_disk_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "attach_disk",
  "arguments": [
    ${vm_cid},
    ${disk_cid}
  ]
}
EOF
  cat ${request_file}

  # Run attach_disk
  printf "%s\n" "Run attach disk method"
  cat ${request_file} | ./rackhd-cpi --configPath=${config_path} 2>&1
}

do_detach_disk() {
  local config_path=$1
  local vm_cid=$2
  local disk_cid=$3

  # Prepare detach disk request
  printf "%s\n" "Prepare detach disk request"
  local request_file="attach_disk_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "detach_disk",
  "arguments": [
    ${vm_cid},
    ${disk_cid}
  ]
}
EOF
  cat ${request_file}

  # Run detach_disk
  printf "%s\n" "Run detach disk method"
  cat ${request_file} | ./rackhd-cpi --configPath=${config_path} 2>&1
}

do_delete_disk() {
  local config_path=$1
  local disk_cid=$2

  # Prepare delete disk request
  printf "%s\n" "Prepare delete disk request"
  local request_file="delete_disk_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "delete_disk",
  "arguments": [
    ${disk_cid}
  ]
}
EOF
  cat ${request_file}

  # Run delete_disk
  printf "%s\n" "Run delete disk method"
  cat ${request_file} | ./rackhd-cpi --configPath=${config_path} 2>&1
}

do_delete_vm() {
  local config_path=$1
  local vm_cid=$2

  # Prepare delete vm request
  printf "%s\n" "Prepare delete vm request"
  local request_file="delete_vm_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "delete_vm",
  "arguments": [
    ${vm_cid}
  ]
}
EOF
  cat ${request_file}

  # Run delete vm method
  printf "%s\n" "Run delete vm method"
  cat ${request_file} | ./rackhd-cpi -configPath=${config_path} 2>&1
}

do_delete_stemcell() {
  local config_path=$1
  local stemcell_id=$2

  # Prepare delete stemcell request
  printf "%s\n" "Prepare delete stemcell request"
  local request_file="delete_stemcell_$(uuidgen)"
  cat > ${request_file} <<EOF
{
  "method": "delete_stemcell",
  "arguments": [
    ${stemcell_id}
  ]
}
EOF
  cat ${request_file}

  # Run delete stemcell method
  printf "%s\n" "Run delete stemcell method"
  cat ${request_file} | ./rackhd-cpi -configPath=${config_path} 2>&1
}

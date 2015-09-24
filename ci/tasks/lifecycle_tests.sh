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
check_param STEMCELL_PATH
check_param CONFIG_PATH
check_param AGENT_ID
check_param PUBLIC_KEY


cd bosh-cpi-release/
source .envrc

bosh_networks=$(cat ci/tasks/bosh_networks_example.json)

go build github.com/onrack/onrack-cpi/onrack-cpi

cat > create_stemcell_request <<EOF
{"method":"create_stemcell", "arguments": ["${STEMCELL_PATH}"]}
EOF

stemcell_id=$(cat create_stemcell_request | ./onrack-cpi -configPath=${CONFIG_PATH} | jq .result)


cat > create_vm_request <<EOF
{"method":"create_vm", "arguments": ["${AGENT_ID}", ${stemcell_id}, {"public_key": "${PUBLIC_KEY}"}, ${bosh_networks}]}
EOF

cat create_vm_request
cat create_vm_request | ./onrack-cpi -configPath=${CONFIG_PATH} 2>&1

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

# bosh-cpi-release/
#source .envrc

go build github.com/onrack/onrack-cpi/onrack-cpi
cd src/github.com/onrack/onrack-cpi/onrack-cpi

cat > create_stemcell_request <<EOF
  {"method":"create_stemcell", "arguments": ["${STEMCELL_PATH}"]}
EOF

stemcell_id=$(cat create_stemcell_request | ./onrack-cpi -configPath=${CONFIG_PATH} | jq .result)

cat > create_vm_request <<EOF
  {"method":"create_vm", "arguments": [${stemcell_id}]}
EOF

cat create_vm_request | ./onrack-cpi

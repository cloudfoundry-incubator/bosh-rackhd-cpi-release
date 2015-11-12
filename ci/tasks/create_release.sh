#!/usr/bin/env bash

set -e -x

pushd ../../
  echo "targeting bosh director at" ${BOSH_VSPHERE_DIRECTOR}
  bosh target ${BOSH_VSPHERE_DIRECTOR}
  expect "Your username:"
  send -- "admin\n"
  expect "Enter password:"
  send -- "admin\n"
popd

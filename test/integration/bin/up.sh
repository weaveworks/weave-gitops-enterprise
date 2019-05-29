#!/bin/bash

export TF_STATE_DIR=$(dirname $0)/../tfstate
mkdir -p $TF_STATE_DIR

$SRCDIR/test/integration/bin/provision_test_vms.sh


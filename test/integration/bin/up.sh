#!/bin/bash

export TF_STATE_DIR=$(dirname $0)/../tfstate
mkdir -p $TF_STATE_DIR

$SRCDIR/test/integration/bin/provision_test_vms.sh

set -e

[ -n "$SECRET_KEY" ] || {
    echo "Cannot run smoke tests: no secret key"
    exit 1
}

# Ensure .ssh directory exists so we can unpack things into it
mkdir -p $HOME/.ssh

# Base name of VMs for integration tests:
export NAME=test-$CIRCLE_BUILD_NUM-$CIRCLE_NODE_INDEX

# Provision and configure testing VMs:
cd "$SRCDIR/test/integration" # Ensures we generate Terraform state files in the right folder, for later use by integration tests.
./bin/run-integration-tests.sh configure
echo "Test VMs now provisioned and configured. $(date)."

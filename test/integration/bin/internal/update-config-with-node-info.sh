#!/bin/bash
#
# Description:
#   This script updates the setup/config.yaml with the
#   GCP node information generated from `run-integration-tests.sh`.
#

set -euo pipefail

unset CD_PATH
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

NUM_MASTERS="${1:-"1"}"
USE_LB="${2:-"false"}"

PRIVATE_HOSTS_FILE="/tmp/hosts_private"
PUBLIC_HOSTS_FILE="/tmp/hosts_public"
SETUP_FILE="config.yaml"

# Need to be inside circle's working dir
cp /tmp/hosts_public /tmp/hosts_private .
jk run \
    -p numMasters=$NUM_MASTERS \
    -p useLB=$USE_LB \
    -p configYamlPath=$SETUP_FILE \
    -i . \
    $SCRIPT_DIR/update-config.js

cat $SETUP_FILE

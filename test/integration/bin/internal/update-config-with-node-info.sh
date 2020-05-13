#!/bin/bash
#
# Description:
#   This script updates the setup/config.yaml with the
#   GCP node information generated from `run-integration-tests.sh`.
#

PRIVATE_HOSTS_FILE="/tmp/hosts_private"
PUBLIC_HOSTS_FILE="/tmp/hosts_public"
SETUP_FILE="config.yaml"

# READ /tmp/hosts_public in array
PUBLIC_IPS=( $(cat /tmp/hosts_public | cut -f 1 -d ' ') )
# READ /tmp/hosts_private in array
PRIVATE_IPS=( $(cat /tmp/hosts_private | cut -f 1 -d ' ') )

# replace all PUBLIC_IP and PRIVATE_IP instances in test/integration/test/config.yaml
# with the corresponding IPs of the GCP nodes in the array
for PUBLIC_IP in ${PUBLIC_IPS[@]}; do
    sed -i "0,/PUBLIC_IP/{s/PUBLIC_IP/$PUBLIC_IP/}" ${SETUP_FILE}
done

for PRIVATE_IP in ${PRIVATE_IPS[@]}; do
    sed -i "0,/PRIVATE_IP/{s/PRIVATE_IP/$PRIVATE_IP/}" ${SETUP_FILE}
done

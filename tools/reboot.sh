#!/bin/bash

# This is a deliberately opinionated script for developing Weave GitOps Enterprise.
# Adapted from the original script in the Weave GitOps repository.
#
# WARN: This script is designed to be "turn it off and on again". It will delete
# the given kind cluster (if it exists) and its GitOps repository and recreate them 
# both, installing everything from scratch.

export KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-wge-dev}"
export GITHUB_REPO="${GITHUB_REPO:-wge-dev}"
export DELETE_GITOPS_DEV_REPO="${DELETE_GITOPS_DEV_REPO:-0}"

do_kind() {
    kind delete cluster --name "$KIND_CLUSTER_NAME"
    kind create cluster --name "$KIND_CLUSTER_NAME" --config "$(dirname "$0")/kind-cluster-with-extramounts.yaml"
}

do_capi(){
    EXP_CLUSTER_RESOURCE_SET=true clusterctl init --infrastructure docker
}

do_flux(){
    if [ "$DELETE_GITOPS_DEV_REPO" == "1" ];
    then
        gh repo delete "$GITHUB_USER/$GITHUB_REPO" --confirm
    fi
    flux bootstrap github --owner="$GITHUB_USER" --repository="$GITHUB_REPO" --branch=main --path=./clusters/management --personal
}

create_local_values_file(){
    envsubst < "$(dirname "$0")/dev-values-local.yaml.tpl" > "$(dirname "$0")/dev-values-local.yaml"
}

main() {
    do_kind
    do_capi
    do_flux
    create_local_values_file
}

main

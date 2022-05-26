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

TOOLS="$(pwd)/tools/bin"

tool_check() {
  if [[ -f "${TOOLS}/${1}" ]]; then
    return
  fi

  echo '!!! Missing tool: '${1}
  echo '    Use "make dependencies" to install all dependencies'

  exit 1
}

do_kind() {
  tool_check "kind"

  ${TOOLS}/kind delete cluster --name "$KIND_CLUSTER_NAME"
  ${TOOLS}/kind create cluster --name "$KIND_CLUSTER_NAME" \
    --config "$(dirname "$0")/kind-cluster-with-extramounts.yaml"
}

do_capi(){
  tool_check "clusterctl"

  EXP_CLUSTER_RESOURCE_SET=true ${TOOLS}/clusterctl init \
    --infrastructure docker
}

do_flux(){
  tool_check "flux"

  if [ "$DELETE_GITOPS_DEV_REPO" == "1" ]; then
    tool_check "gh"

    ${TOOLS}/gh repo delete "$GITHUB_USER/$GITHUB_REPO" --confirm
  fi

  ${TOOLS}/flux bootstrap github \
    --owner="$GITHUB_USER" \
    --repository="$GITHUB_REPO" \
    --branch=main \
    --path=./clusters/management \
    --personal
}

create_local_values_file(){
  envsubst \
    < "$(dirname "$0")/dev-values-local.yaml.tpl" \
    > "$(dirname "$0")/dev-values-local.yaml"
}

main() {
  do_kind
  do_capi
  do_flux
  create_local_values_file
}

main

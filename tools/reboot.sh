#!/bin/bash

# This is a deliberately opinionated script for developing Weave GitOps Enterprise.
# Adapted from the original script in the Weave GitOps repository.
#
# WARN: This script is designed to be "turn it off and on again". It will delete
# the given kind cluster (if it exists) and its GitOps repository and recreate them 
# both, installing everything from scratch.

source $PWD/tools/flags.env

do_kind() {
  tool_check "kind"
  tool_check "yq"

  if [ -n "$(ls "$(dirname "$0")"/custom/kind-cluster-patch-*.yaml 2> /dev/null)" ] ; then
      "${TOOLS}"/yq eval-all '. as $item ireduce ({}; . *d $item)' "$(dirname "$0")"/kind-cluster-with-extramounts.yaml "$(dirname "$0")"/custom/kind-cluster-patch-*.yaml > "$(dirname "$0")"/kind-config.yaml
  else
      cp "$(dirname "$0")"/kind-cluster-with-extramounts.yaml "$(dirname "$0")"/kind-config.yaml
  fi

  ${TOOLS}/kind delete cluster --name "$KIND_CLUSTER_NAME"
  ${TOOLS}/kind create cluster --name "$KIND_CLUSTER_NAME" \
    --config "$(dirname "$0")/kind-config.yaml"
}

main() {
  github_env_check
  do_kind
  $PWD/tools/setup.sh
}

main

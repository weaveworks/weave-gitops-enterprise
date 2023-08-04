#!/usr/bin/env bash


# figure out a basepath relative to the location of this script
# so you can run this script from any path
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
args=("$@")
CLUSTER_NAME=${args[1]:-kind-smoke}

set -x 

function setup_kind {
  kind create cluster --name $CLUSTER_NAME --config ${SCRIPT_DIR}/../data/kind/local-kind-config.yaml
  kubectl wait --for=condition=Ready --timeout=120s -n kube-system pods --all
  kubectl get pods -A
  exit 0
}

setup_kind

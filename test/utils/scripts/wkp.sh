#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'setup' ])
then 
    echo "Invalid option, valid values => [ setup ]"
    exit 1
fi

set -x 

function setup {
  if [ ${#args[@]} -ne 3 ]
  then
    echo "Cluster name and workspace path both are required arguments"
    exit 1
  fi

  if [ "$RUNNER_OS" == "macOS" ]; then
    SED_BIN="gsed"
  else
    SED_BIN="sed"
  fi

  export SKIP_PROMPT=1
  export CLUSTER_NAME=${args[1]}
  
  wkp_kind_cluster_dir=$(mktemp -d)
  cd $wkp_kind_cluster_dir
  $WKP_BIN_PATH setup install --entitlements=${args[2]}/test/ci-wks-unlimited.entitlements
  echo "${DOCKER_IO_PASSWORD}" > /tmp/docker-io-password
  $SED_BIN -i "s/track:.*/track: wks-components/" ${wkp_kind_cluster_dir}/setup/config.yaml
  $SED_BIN -i "s/clusterName:.*/clusterName: ${CLUSTER_NAME}/" ${wkp_kind_cluster_dir}/setup/config.yaml
  $SED_BIN -i "s/gitProvider:.*/gitProvider: github/" ${wkp_kind_cluster_dir}/setup/config.yaml
  $SED_BIN -i "s/gitProviderOrg:.*/gitProviderOrg: ${GITHUB_ORG}/" ${wkp_kind_cluster_dir}/setup/config.yaml
  $SED_BIN -i "s/dockerIOUser:.*/dockerIOUser: ${DOCKER_IO_USER}/" ${wkp_kind_cluster_dir}/setup/config.yaml 
  $SED_BIN -i "s|dockerIOPasswordFile:.*|dockerIOPasswordFile: /tmp/docker-io-password|" ${wkp_kind_cluster_dir}/setup/config.yaml 
  # don't skip the components we need those
  export SKIP_COMPONENTS=false
  $WKP_BIN_PATH setup run --entitlements=${args[2]}/test/ci-wks-unlimited.entitlements
  # Wait until pods are running
  kubectl wait nodes --all --for=condition=ready --timeout=120s || true
  kubectl wait pods -n kube-system -l tier=control-plane --for=condition=ready --timeout=120s || true
  kubectl wait deployment.apps/coredns -n kube-system --for=condition=available --timeout=120s || true
  
  kubectl get pods -A
  kubectl get nodes
  $WKP_BIN_PATH workspaces add-provider \
      --type github \
      --token "${GITHUB_TOKEN}" \
      --secret-name github-token \
      --git-commit-push \
      --entitlements=${args[2]}/test/ci-wks-unlimited.entitlements
}

if [ ${args[0]} = 'setup' ]; then
    setup
fi

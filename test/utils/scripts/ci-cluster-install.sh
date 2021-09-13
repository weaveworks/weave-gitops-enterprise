#!/usr/bin/env bash

args=("$@")

if [ -z ${args[0]} ] || ([ ${args[0]} != 'mccp' ] && [ ${args[0]} != 'wkp' ])
then 
    echo "Invalid option, valid values => [ mccp, wkp ]"
    exit 1
fi

set -x 

function setup_mccp {
  if [ ${#args[@]} -ne 2 ]
  then
    echo "Cluster repository name is required arguments for 'mccp' setup."
    exit 1
  fi

  # create unique cluster config repository name
  CLUSTER_REPOSITORY=${args[1]}
  GIT_REPOSITORY_URL="https://github.com/$GITHUB_ORG/$CLUSTER_REPOSITORY"

  # Set the CLUSTER_REPOSITORY as environment variable for subsequent steps
  echo "CLUSTER_REPOSITORY=$CLUSTER_REPOSITORY" >> $GITHUB_ENV

  WORKER_NODE=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name)
  kubectl label "${WORKER_NODE}" wkp-database-volume-node=true

  if [ "$RUNNER_OS" == "Linux" ]; then
    WORKER_NODE_EXTERNAL_IP=$(ifconfig eth0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:)
  elif [ "$RUNNER_OS" == "macOS" ]; then
    WORKER_NODE_EXTERNAL_IP=$(ifconfig eth0 | grep -i MASK | awk '{print $2}' | cut -f2 -d:)
  fi

  echo "Worker node ip is ${WORKER_NODE_EXTERNAL_IP}"
  NATS_NODEPORT=31490            

  kubectl create namespace prom
  helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
  helm repo update
  helm install my-prom prometheus-community/kube-prometheus-stack \
    --namespace prom \
    --version 14.4.0 \
    --values test/utils/data/mccp-prometheus-values.yaml

  kubectl create namespace mccp
  kubectl create secret docker-registry docker-io-pull-secret \
    --namespace mccp \
    --docker-username="${DOCKER_IO_USER}" \
    --docker-password="${DOCKER_IO_PASSWORD}" 
  kubectl create secret generic git-provider-credentials \
    --namespace=mccp \
    --from-literal="GIT_PROVIDER_TOKEN=${GITHUB_TOKEN}"
  CHART_VERSION=$(git describe --always | sed 's/^[^0-9]*//')
  helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3/
  helm repo update
  helm install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace mccp \
    --set "imagePullSecrets[0].name=docker-io-pull-secret" \
    --set "wkp-ui.image.pullSecrets[0]=docker-io-pull-secret" \
    --set "nats.client.service.nodePort=${NATS_NODEPORT}" \
    --set "agentTemplate.natsURL=${WORKER_NODE_EXTERNAL_IP}:${NATS_NODEPORT}" \
    --set "config.capi.repositoryURL=${GIT_REPOSITORY_URL}" \
    --set "config.capi.baseBranch=main"

  # Wait for cluster to settle
  kubectl wait --for=condition=Ready --timeout=300s -n mccp --all pod
  kubectl get pods -A
}

function setup_wkp {
  if [ ${#args[@]} -ne 3 ]
  then
    echo "Cluster name and workspace path both are required arguments for 'wkp' setup."
    exit 1
  fi
  export SKIP_PROMPT=1
  export CLUSTER_NAME=${args[1]}
  
  wkp_kind_cluster_dir=$(mktemp -d)
  cd $wkp_kind_cluster_dir
  $WKP_BIN_PATH setup install --entitlements=${args[2]}/test/ci-wks-unlimited.entitlements
  echo "${DOCKER_IO_PASSWORD}" > /tmp/docker-io-password
  sed -i "s/track:.*/track: wks-components/" ${wkp_kind_cluster_dir}/setup/config.yaml
  sed -i "s/clusterName:.*/clusterName: ${CLUSTER_NAME}/" ${wkp_kind_cluster_dir}/setup/config.yaml
  sed -i "s/gitProvider:.*/gitProvider: github/" ${wkp_kind_cluster_dir}/setup/config.yaml
  sed -i "s/gitProviderOrg:.*/gitProviderOrg: ${GITHUB_ORG}/" ${wkp_kind_cluster_dir}/setup/config.yaml
  sed -i "s/dockerIOUser:.*/dockerIOUser: ${DOCKER_IO_USER}/" ${wkp_kind_cluster_dir}/setup/config.yaml 
  sed -i "s|dockerIOPasswordFile:.*|dockerIOPasswordFile: /tmp/docker-io-password|" ${wkp_kind_cluster_dir}/setup/config.yaml 
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

if [ ${args[0]} = 'mccp' ]
then
    setup_mccp
fi

if [ ${args[0]} = 'wkp' ]
then
    setup_wkp
fi

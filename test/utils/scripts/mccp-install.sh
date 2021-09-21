#!/usr/bin/env bash

UI_NODEPORT=30080
NATS_NODEPORT=31490
WORKER_NODE_EXTERNAL_IP=$(ipconfig getifaddr en0)

WORKER_NODE=$(kubectl get node --selector='!node-role.kubernetes.io/master' -o name | head -n 1)
kubectl label "${WORKER_NODE}" wkp-database-volume-node=true

kubectl create namespace prom
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install my-prom prometheus-community/kube-prometheus-stack \
  --namespace prom \
  --version 14.4.0 \
  --values test/utils/data/mccp-prometheus-values.yaml

kubectl create namespace wego-system
kubectl apply -f test/utils/scripts/entitlement-secret.yaml
kubectl create secret docker-registry docker-io-pull-secret \
  --namespace wego-system \
  --docker-username="${DOCKER_IO_USER}" \
  --docker-password="${DOCKER_IO_PASSWORD}"
kubectl create secret generic git-provider-credentials \
  --namespace=wego-system \
  --from-literal="GIT_PROVIDER_TOKEN=${GITHUB_TOKEN}"
# CHART_VERSION=$(git describe --always | sed 's/^[^0-9]*//')
CHART_VERSION=0.0.7-27-g707be80
helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/releases/charts-v3/
helm repo update
helm upgrade --install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace wego-system \
  --set "imagePullSecrets[0].name=docker-io-pull-secret" \
  --set "wkp-ui.image.pullSecrets[0]=docker-io-pull-secret" \
  --set "nats.client.service.nodePort=${NATS_NODEPORT}" \
  --set "agentTemplate.natsURL=${WORKER_NODE_EXTERNAL_IP}:${NATS_NODEPORT}" \
  --set "nginx-ingress-controller.service.nodePorts.http=${UI_NODEPORT}" \
  --set "nginx-ingress-controller.service.type=NodePort" \
  --set "config.capi.repositoryURL=${GIT_REPOSITORY_URL}" \
  --set "config.capi.baseBranch=main"
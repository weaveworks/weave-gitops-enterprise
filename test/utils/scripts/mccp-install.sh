#!/usr/bin/env bash

UI_NODEPORT=30080
NATS_NODEPORT=31490
WORKER_NODE_EXTERNAL_IP=$(ipconfig getifaddr en0)

kubectl create namespace prom
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install my-prom prometheus-community/kube-prometheus-stack \
  --namespace prom \
  --version 14.4.0 \
  --values test/utils/data/mccp-prometheus-values.yaml

kubectl create namespace mccp
kubectl create secret docker-registry docker-io-pull-secret \
  --namespace mccp \
  --docker-username="${DOCKER_IO_USER}" \
  --docker-password="${DOCKER_IO_PASSWORD}"
# FIXME: should be pattern matched again weave-gitops tags
CHART_VERSION=$(git describe --always --match "weave-gitops*" --abbrev=8 HEAD| sed 's/^[^0-9]*//')
helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3/
helm repo update
helm upgrade --install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace mccp \
  --set "imagePullSecrets[0].name=docker-io-pull-secret" \
  --set "wkp-ui.image.pullSecrets[0]=docker-io-pull-secret" \
  --set "nats.client.service.nodePort=${NATS_NODEPORT}" \
  --set "agentTemplate.natsURL=${WORKER_NODE_EXTERNAL_IP}:${NATS_NODEPORT}" \
  --set "nginx-ingress-controller.service.nodePorts.http=${UI_NODEPORT}"

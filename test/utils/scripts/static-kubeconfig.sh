#!/bin/bash

if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Ensure CLUSTER_NAME has been set"
    exit 1
fi

if [[ -z "$CA_CERTIFICATE" ]]; then
    echo "Ensure CA_CERTIFICATE has been set to the path of the CA certificate"
    exit 1
fi

if [[ -z "$ENDPOINT" ]]; then
    echo "Ensure ENDPOINT has been set"
    exit 1
fi

if [[ -z "$TOKEN" ]]; then
    echo "Ensure TOKEN has been set"
    exit 1
fi

export CLUSTER_CA_CERTIFICATE=$(cat "$CA_CERTIFICATE" | base64)

envsubst <<EOF
apiVersion: v1
kind: Config
clusters:
- name: $CLUSTER_NAME
  cluster:
    server: https://$ENDPOINT
    certificate-authority-data: $CLUSTER_CA_CERTIFICATE
users:
- name: $CLUSTER_NAME
  user:
    token: $TOKEN
contexts:
- name: $CLUSTER_NAME
  context:
    cluster: $CLUSTER_NAME
    user: $CLUSTER_NAME
current-context: $CLUSTER_NAME

EOF
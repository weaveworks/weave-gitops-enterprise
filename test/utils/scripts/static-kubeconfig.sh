#!/bin/bash

if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Ensure CLUSTER_NAME has been set"
    exit 1
fi

if [[ -z "$CA_AUTHORITY" ]]; then
    echo "Ensure CA_AUTHORITY has been set"
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

envsubst <<EOF
apiVersion: v1
kind: Config
clusters:
- name: $CLUSTER_NAME
  cluster:
    server: $ENDPOINT
    certificate-authority-data: $CA_AUTHORITY
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
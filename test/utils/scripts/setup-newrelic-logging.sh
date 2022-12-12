#!/bin/bash

if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Ensure CLUSTER_NAME has been set"
    exit 1
fi

if [[ -z "$NEWRELIC_LICENSE_KEY" ]]; then
    echo "Ensure NEWRELIC_LICENSE_KEY has been set"
    exit 1
fi

# Setup values.yaml for New Relic helm chart
envsubst <<EOF > values.yaml
global:
    licenseKey: $NEWRELIC_LICENSE_KEY
    cluster: $CLUSTER_NAME
endpoint: https://log-api.eu.newrelic.com/log/v1

EOF


flux create source helm newrelic \
    --namespace flux-system \
    --url="https://helm-charts.newrelic.com" \
    --interval=1m0s

flux create helmrelease newrelic-logging \
    --namespace=flux-system \
    --interval=1m0s \
    --source=HelmRepository/newrelic \
    --chart=newrelic-logging \
    --chart-version="1.12.0" \
    --values=./values.yaml

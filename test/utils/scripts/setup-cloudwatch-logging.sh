#!/bin/bash

if [[ -z "$CLUSTER_NAME" ]]; then
    echo "Ensure CLUSTER_NAME has been set"
    exit 1
fi

if [[ -z "$CLOUDWATCH_AWS_ACCESS_KEY_ID" ]]; then
    echo "Ensure CLOUDWATCH_AWS_ACCESS_KEY_ID has been set"
    exit 1
fi

if [[ -z "$CLOUDWATCH_AWS_SECRET_ACCESS_KEY" ]]; then
    echo "Ensure CLOUDWATCH_AWS_SECRET_ACCESS_KEY has been set"
    exit 1
fi

# Generate values.yaml for aws-for-fluent-bit helm chart
envsubst <<EOF > values.yaml
cloudWatch:
  enabled: true
  region: "eu-north-1"
  logGroupName: "/fluentbit/logs/$CLUSTER_NAME"
firehose:
  enabled: false
kinesis:
  enabled: false
elasticsearch:
  enabled: false
env:
  - name: AWS_ACCESS_KEY_ID
    value: $CLOUDWATCH_AWS_ACCESS_KEY_ID
  - name: AWS_SECRET_ACCESS_KEY
    value: $CLOUDWATCH_AWS_SECRET_ACCESS_KEY

EOF


flux create source helm eks \
    --namespace flux-system \
    --url="https://aws.github.io/eks-charts" \
    --interval=1m0s

flux create helmrelease aws-fluent-bit \
    --namespace=flux-system \
    --interval=1m0s \
    --source=HelmRepository/eks \
    --chart=aws-for-fluent-bit \
    --chart-version="0.1.21" \
    --values=./values.yaml

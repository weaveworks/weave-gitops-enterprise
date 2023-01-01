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

helm install aws-for-fluent-bit eks/aws-for-fluent-bit \
    --namespace kube-system \
    --values=./values.yaml

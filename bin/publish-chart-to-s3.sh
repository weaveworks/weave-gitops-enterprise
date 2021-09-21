#!/usr/bin/env bash

# Test on OSX w/:
# SED=gsed ./tools/publish-chart-to-s3.sh v0.1.2

# Requires Helm v2 (helm2) and Helm v3 (helm) to be installed. Defaults to Helm v3.
# Also requires aws cli for uploading to s3.

set -x

set -o errexit
set -o pipefail

TAG=$1
PUBLISH_EVENT=$2
CHART=$3
CHART_VERSION=$4
SED=${SED:-"sed"}
HELM=${HELM:-"helm"}

$HELM version --client

if [ "${CHART_VERSION}" == "2" ]; then
    # Init
    $HELM init --client-only
fi
SEMVER=$(echo $TAG | $SED 's/^[^0-9]*//')
$SED -i "s/^version: .*$/version: $SEMVER/" ${CHART}/Chart.yaml
$HELM lint "${CHART}"

rm -rf ./pkg
mkdir -p ./pkg
$HELM package "${CHART}" --dependency-update --destination ./pkg/

cd pkg
if [ "${CHART_VERSION}" == "2" ]; then
    # Download the existing index.yaml from s3 and update it
    aws s3 cp s3://weaveworks-wkp/charts/index.yaml . || echo "No index.yaml found..."
    $HELM repo index . --merge index.yaml --url https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts/
    # Upload index and chart to s3
    aws s3 cp index.yaml s3://weaveworks-wkp/charts/
    aws s3 cp *.tgz s3://weaveworks-wkp/charts/
elif [ "${CHART_VERSION}" == "3" ]; then
    if [ "${PUBLISH_EVENT}" == "release" ]; then
        URI="releases/charts-v3"
    elif [ "${PUBLISH_EVENT}" == "schedule" ]; then
        URI="nightly/charts-v3"
    else
        URI="charts-v3"
    fi
    # Download the existing index.yaml from s3 and update it
    aws s3 cp s3://weaveworks-wkp/${URI}/index.yaml . || echo "No index.yaml found..."
    $HELM repo index . --merge index.yaml --url https://s3.us-east-1.amazonaws.com/weaveworks-wkp/${URI}/
    # Upload index and chart to s3
    aws s3 cp index.yaml s3://weaveworks-wkp/${URI}/
    aws s3 cp *.tgz s3://weaveworks-wkp/${URI}/
else
    echo "Helm version can be '2' or '3'"
fi

# CAREFUL, this is done by .circleci but you can do it manually here if you need to.
# Upload index and new release .tgz back to s3
# aws s3 cp pkg/index.yaml s3://weaveworks-wkp/charts/
# aws s3 cp pkg/*.tgz s3://weaveworks-wkp/charts/

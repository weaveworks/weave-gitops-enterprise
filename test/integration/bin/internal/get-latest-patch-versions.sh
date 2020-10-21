#!/bin/bash
#
# Description:
#   This script generates the environment variables to
#   run the integration upgrade tests for the latest available
#   patch versions given a specific minor version of K8s.
#   e.g. for minor versions 1.14,1.15,1.16,1.17 it should return:
#   UPGRADE_VERSIONS='1.14.10,1.15.12,1.16.13,1.17.9'
#   as at the time of writing, the above are the latest patch versions
#

set -euo pipefail

KUBECTL_DOWNLOAD_URL='https://storage.googleapis.com/kubernetes-release/release/v%s/bin/linux/amd64/kubectl'

get_version_status_code() {
    RENDERED_URL=$(printf ${KUBECTL_DOWNLOAD_URL} $1)
    echo "$(curl -o /dev/null --silent --head --write-out '%{http_code}\n' ${RENDERED_URL})"
}

LAST_PATCH_VERSION=""
UPGRADE_VERSIONS=$1

IFS=','; set -f
MINOR_VERSIONS=($2)
for MINOR_VERSION in "${MINOR_VERSIONS[@]}"
do
    for PATCH_VERSION in {0..99}
    do
        VERSION_STRING="${MINOR_VERSION}.${PATCH_VERSION}"
        STATUS_CODE=$(get_version_status_code "${VERSION_STRING}" 2>&1)
        if [[ ${STATUS_CODE} = "200" ]]; then
            LAST_PATCH_VERSION=${VERSION_STRING}
        else
            UPGRADE_VERSIONS+=",${LAST_PATCH_VERSION}"
            break
        fi
    done
done

echo "export UPGRADE_VERSIONS='${UPGRADE_VERSIONS}'"

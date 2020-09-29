#!/usr/bin/env sh

job="${1}"
branch="${2:-"$(git rev-parse --abbrev-ref HEAD)"}"

if [ -z "${job}" ]; then
    echo "Specify a circle job! e.g.:"
    echo ""
    echo "./tools/trigger-single-integration-test.sh cluster-components-gcp"
    echo ""
    exit 1
fi

if [ -z "${CIRCLECI_TOKEN}" ]; then
    echo "No CIRCLECI_TOKEN env var set, run w/:"
    echo ""
    echo "CIRCLECI_TOKEN=\$MY_PERSONAL_CIRCLECI_TOKEN ./tools/trigger-single-integration-test.sh cluster-components-gcp"
    echo ""
    exit 1
fi

git fetch
tag=$(./tools/image-tag "${branch}")
url="https://s3.amazonaws.com/weaveworks-wkp/wk-${tag}-linux-amd64"
curl --fail --head "${url}"
res=$?
if [ "${res}" != 0 ]; then
    echo "Binary ${url} doesn't seem to exist so this test won't work..."
    echo "Maybe its still being built?"
    exit 1
fi

echo "Testing ${job} on branch ${branch}... (${tag})"

curl --fail \
    -u ${CIRCLECI_TOKEN}: \
    -d "build_parameters[CIRCLE_JOB]=${job}" \
    "https://circleci.com/api/v1.1/project/github/weaveworks/wks/tree/${branch}"


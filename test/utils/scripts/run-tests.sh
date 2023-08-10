#!/usr/bin/env bash

args=("$@")

set -x 
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
WORKSPACE_PATH=$(dirname $(dirname $(dirname ${SCRIPT_DIR})))

export GIT_PROVIDER=${GIT_PROVIDER:-gitlab}
export GIT_PROVIDER_HOSTNAME=${GIT_PROVIDER_HOSTNAME:-gitlab.git.dev.weave.works}
export GITLAB_ORG=${GITLAB_ORG:-${GITLAB_USER}-org}
export GITLAB_CLIENT_ID="438bb793d4815349394735dad8644406d5f9ffd7b8d861ef61984d1cbee7df3c"
export GITLAB_CLIENT_SECRET="e3c613dab49ebd7d4d921fe60f4c649a8c414497fd54181b541c6f95f1b3a66d"
export WEAVE_GITOPS_GIT_HOST_TYPES="gitlab.git.dev.weave.works=gitlab"
export GITLAB_HOSTNAME="gitlab.git.dev.weave.works"
export CLUSTER_REPOSITORY=${CLUSTER_REPOSITORY:-smoke-tests}
export OIDC_ISSUER_URL=${OIDC_ISSUER_URL:-https://dex-01.wge.dev.weave.works}
export DEX_CLIENT_ID=${DEX_CLIENT_ID:-weave-gitops-enterprise}
export DEX_CLIENT_SECRET=${DEX_CLIENT_SECRET:-2JPIcb5IvO1isJ3Zii7jvjqbUtLtTC}
export UI_NODEPORT=${UI_NODEPORT:-30080}

export GITOPS_BIN_PATH=`which gitops`
# export LOGIN_USER_TYPE="cluster-user"
export LOGIN_USER_TYPE="oidc"

ginkgo --label-filter='smoke&&tenant' --v --output-dir=/tmp/foot-smoke --timeout=2h ${WORKSPACE_PATH}/test/acceptance/test/
# ginkgo --label-filter='smoke&&capd' --v --output-dir=/tmp/foot-smoke --timeout=2h ${WORKSPACE_PATH}/test/acceptance/test/

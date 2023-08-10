#!/usr/bin/env bash

args=("$@")

set -x 
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
WORKSPACE_PATH=$(dirname $(dirname $(dirname ${SCRIPT_DIR})))

GIT_PROVIDER=${GIT_PROVIDER:-gitlab}
GIT_PROVIDER_HOSTNAME=${GIT_PROVIDER_HOSTNAME:-gitlab.git.dev.weave.works}
GITLAB_ORG=${GITLAB_ORG:-${GITLAB_USER}-org}
GITLAB_CLIENT_ID="438bb793d4815349394735dad8644406d5f9ffd7b8d861ef61984d1cbee7df3c"
GITLAB_CLIENT_SECRET="e3c613dab49ebd7d4d921fe60f4c649a8c414497fd54181b541c6f95f1b3a66d"
WEAVE_GITOPS_GIT_HOST_TYPES="gitlab.git.dev.weave.works=gitlab"
GITLAB_HOSTNAME="gitlab.git.dev.weave.works"
CLUSTER_REPOSITORY=${CLUSTER_REPOSITORY:-smoke-tests}
OIDC_ISSUER_URL=${OIDC_ISSUER_URL:-https://dex-01.wge.dev.weave.works}
DEX_CLIENT_ID=${DEX_CLIENT_ID:-weave-gitops-enterprise}
DEX_CLIENT_SECRET=${DEX_CLIENT_SECRET:-2JPIcb5IvO1isJ3Zii7jvjqbUtLtTC}
UI_NODEPORT=${UI_NODEPORT:-30080}

function preflight {
  # check that required env vars are set
  if [ -z ${GITHUB_TOKEN} ] && [ -z ${GITLAB_TOKEN} ]; then
    echo "Please set GITHUB_TOKEN or GITLAB_TOKEN"
    exit 1
  fi

  # if GIT_PROVIDER is gitlab then check the user, pass and token vars are provided
  if [ ${GIT_PROVIDER} == "gitlab" ]; then
    if [ -z ${GITLAB_USER} ] || [ -z ${GITLAB_PASSWORD} ] || [ -z ${GITLAB_TOKEN} ]; then
      echo "Please set GITLAB_USER, GITLAB_PASSWORD and GITLAB_TOKEN"
      exit 1
    fi
  fi
}

function setup {
  helm repo add wkpv3 https://s3.us-east-1.amazonaws.com/weaveworks-wkp/charts-v3-r2/
  helm repo update  
  
  flux install

  # Create secrete for git provider authentication
  if [ ${GIT_PROVIDER} == "github" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITHUB_ORG/$CLUSTER_REPOSITORY.git

    flux bootstrap github \
      --owner=${GITHUB_ORG} \
      --repository=${CLUSTER_REPOSITORY} \
      --branch=main \
      --path=./clusters/management \
      --interval=30s

  elif [ ${GIT_PROVIDER} == "gitlab" ]; then
    GIT_REPOSITORY_URL="https://$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY"
    GITOPS_REPO=ssh://git@$GIT_PROVIDER_HOSTNAME/$GITLAB_ORG/$CLUSTER_REPOSITORY.git

    # if WEAVE_GITOPS_GIT_HOST_TYPES is set, create secret for git provider authentication
    if [ ! -z ${WEAVE_GITOPS_GIT_HOST_TYPES} ]; then
      # delete secret if present then recreate
      kubectl delete secret git-provider-credentials --namespace=flux-system || echo "secret not found"
      kubectl create secret generic git-provider-credentials --namespace=flux-system \
      --from-literal="GITLAB_CLIENT_ID=$GITLAB_CLIENT_ID" \
      --from-literal="GITLAB_CLIENT_SECRET=$GITLAB_CLIENT_SECRET" \
      --from-literal="GITLAB_HOSTNAME=$GIT_PROVIDER_HOSTNAME" \
      --from-literal="GIT_HOST_TYPES=$WEAVE_GITOPS_GIT_HOST_TYPES"
    fi

    flux bootstrap gitlab \
      --owner=${GITLAB_ORG} \
      --repository=${CLUSTER_REPOSITORY} \
      --branch=main \
      --hostname=${GIT_PROVIDER_HOSTNAME} \
      --path=./clusters/management \
      --interval=30s
  fi  

  # kubectl wait --for=condition=Ready --timeout=300s -n flux-system --all pod
    
  # Create admin cluster user secret
  kubectl apply -f ${WORKSPACE_PATH}/test/utils/data/auth/base.yaml
  
  #  Create client credential secret for OIDC (dex)
  kubectl delete secret --namespace flux-system client-credentials || echo "secret not found"
  kubectl create secret generic client-credentials \
  --namespace flux-system \
  --from-literal=clientID=${DEX_CLIENT_ID} \
  --from-literal=clientSecret=${DEX_CLIENT_SECRET}

  kubectl apply -f ${WORKSPACE_PATH}/test/utils/data/entitlement/entitlement-secret.yaml 

  # Choosing weave-gitops-enterprise chart version to install
  if [ -z ${ENTERPRISE_CHART_VERSION} ]; then
    CHART_VERSION=$(git describe --always --abbrev=7 | sed 's/^[^0-9]*//')
  else
    CHART_VERSION=${ENTERPRISE_CHART_VERSION}
  fi

  # enable cluster resource sets
  export EXP_CLUSTER_RESOURCE_SET=true
  # Install capi infrastructure provider
  clusterctl init --infrastructure docker   
  kubectl wait --for=condition=Ready --timeout=300s -n capd-system --all pod 

  # Install weave gitops enterprise controllers
  helmArgs=()
  helmArgs+=( --set "service.type=NodePort" )
  helmArgs+=( --set "service.ports.https=8000" )
  helmArgs+=( --set "service.nodePorts.https=30080" )
  helmArgs+=( --set "config.git.type=${GIT_PROVIDER}" )
  helmArgs+=( --set "config.git.hostname=${GIT_PROVIDER_HOSTNAME}" )
  helmArgs+=( --set "config.capi.repositoryURL=${GIT_REPOSITORY_URL}" )
  # using default repository path '"./clusters/management/clusters"' so the application reconciliation always happen out of the box
  # helmArgs+=( --set "config.capi.repositoryPath=./clusters/my-cluster/clusters" )
  helmArgs+=( --set "config.capi.repositoryClustersPath=./clusters" )
  helmArgs+=( --set "config.capi.baseBranch=main" )
  helmArgs+=( --set "tls.enabled=false" )
  helmArgs+=( --set "config.oidc.enabled=true" )
  helmArgs+=( --set "config.oidc.clientCredentialsSecret=client-credentials" )
  helmArgs+=( --set "config.oidc.issuerURL=${OIDC_ISSUER_URL}" )
  helmArgs+=( --set "config.oidc.redirectURL=http://localhost:${UI_NODEPORT}/oauth2/callback" )
  helmArgs+=( --set "policy-agent.enabled=true" )
  helmArgs+=( --set "policy-agent.config.accountId=weaveworks" )
  helmArgs+=( --set "policy-agent.config.clusterId=management" )
  helmArgs+=( --set "features.progressiveDelivery.enabled=true" )
  # helmArgs+=( --set "extraEnvVars[0].name=WEAVE_GITOPS_USERS_CLIENTS_TTL" )
  # helmArgs+=( --set-string "extraEnvVars[0].value=5m" )

 
  helm upgrade --install my-mccp wkpv3/mccp --version "${CHART_VERSION}" --namespace flux-system --wait ${helmArgs[@]}
  
   # Wait for cluster to settle
  kubectl wait --for=condition=Ready --timeout=300s -n flux-system --all pod

  # Create profiles HelmReposiotry 'weaveworks-charts'
  flux create source helm weaveworks-charts --url="https://raw.githubusercontent.com/weaveworks/profiles-catalog/gh-pages" --interval=30s --namespace flux-system 

  # Install RBAC for user authentication
  kubectl apply -f ${WORKSPACE_PATH}/test/utils/data/rbac/user-role-bindings.yaml

  kubectl get pods -A

  exit 0
}

preflight
setup


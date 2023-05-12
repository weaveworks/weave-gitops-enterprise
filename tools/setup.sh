#!/bin/bash

# This paves a cluster with the prerequisites for Weave GitOps
# Enterprise development:
# - a stock Flux installation, bootstrapped against the git repo given
#   in GITHUB_USER and GITHUB_REPO
# - a CAPI installation that includes the vcluster provider
#
# These are (mostly) idempotent -- calling this script repeatedly
# won't break anything. You will probably need to rm -rf
# /tmp/$GITHUB_REPO before running, though, because the script plays
# it safe with deleting that.
#
# This script will get called by ./reboot.sh after recreating the kind
# cluster, and you can call it separately if you don't want to
# recreate the kind cluster first. This might be useful if ./reboot.sh
# repeatedly fails to make progress.

set -euo pipefail

source $PWD/tools/flags.env

do_capi(){
  tool_check "clusterctl"

  EXP_CLUSTER_RESOURCE_SET=true ${TOOLS}/clusterctl init \
    --infrastructure vcluster
}

do_flux(){
  tool_check "flux"

  if [ "$DELETE_GITOPS_DEV_REPO" == "1" ]; then
    tool_check "gh"

    ${TOOLS}/gh repo delete "$GITHUB_USER/$GITHUB_REPO" --yes
  fi

  ${TOOLS}/flux bootstrap github \
    --owner="$GITHUB_USER" \
    --repository="$GITHUB_REPO" \
    --branch=main \
    --path=./clusters/management \
    --personal
}

create_local_values_file(){
  envsubst \
    < "$(dirname "$0")/dev-values-local.yaml.tpl" \
    > "$(dirname "$0")/dev-values-local.yaml"
}

add_files_to_git(){
  tool_check "gh"
  # We could use $GITHUB_REPO here, but its rm -rf so we'll be careful
  rm -rf "/tmp/wge-dev"
  ${TOOLS}/gh repo clone "ssh://git@github.com/$GITHUB_USER/$GITHUB_REPO" "/tmp/$GITHUB_REPO"
  mkdir -p "/tmp/$GITHUB_REPO/clusters/bases/rbac"
  mkdir -p "/tmp/$GITHUB_REPO/clusters/bases/networkpolicy"
  cp "$(dirname "$0")/git-files/wego-admin.yaml" "/tmp/$GITHUB_REPO/clusters/bases/rbac/wego-admin.yaml"
  cp "$(dirname "$0")/git-files/flux-system-networkpolicy.yaml" "/tmp/$GITHUB_REPO/clusters/bases/networkpolicy/flux-system-networkpolicy.yaml"
  pushd "/tmp/$GITHUB_REPO"
  git add clusters/bases
  git commit -m "Add wego-admin role"
  git push origin main
  popd
}

# Steps we ask you to do in https://docs.gitops.weave.works/docs/cluster-management/getting-started/
follow_capi_user_guide(){
  add_files_to_git
  kubectl create secret generic my-pat --from-literal GITHUB_TOKEN="$GITHUB_TOKEN" --from-literal GITHUB_USER="$GITHUB_USER" --from-literal GITHUB_REPO="$GITHUB_REPO"
}

push_progressive_delivery_manifests_to_gitops_dev_repo(){
  if [ "$PUSH_PROGRESSIVE_DELIVERY_MANIFESTS_TO_GITOPS_DEV_REPO" == "1" ]; then
    if [ ! -d "$(dirname "$0")/../../progressive-delivery" ]; then
      echo '!!! Missing directory: '${1}
      echo '    Ensure the "weaveworks/progressive-delivery" repository is checked out in a directory that is adjacent to this repository.'

      exit 1
    fi

    tool_check "gh"

    # We could use $GITHUB_REPO here, but its rm -rf so we'll be careful
    rm -rf "/tmp/wge-dev"
    ${TOOLS}/gh repo clone "ssh://git@github.com/$GITHUB_USER/$GITHUB_REPO" "/tmp/$GITHUB_REPO"
    mkdir -p "/tmp/$GITHUB_REPO/apps/progressive-delivery"
    cp -r "$(dirname "$0")/../../progressive-delivery/tools/extra-resources/" "/tmp/$GITHUB_REPO/apps/progressive-delivery/"
    rm "/tmp/$GITHUB_REPO/apps/progressive-delivery/istio/resources/gateway.yaml"
    cp "$(dirname "$0")/git-files/progressive-delivery-kustomizations.yaml" "/tmp/$GITHUB_REPO/clusters/management/progressive-delivery-kustomizations.yaml"
    pushd "/tmp/$GITHUB_REPO"
    git add apps/progressive-delivery
    git add clusters/management
    git commit -m "Add progressive-delivery manifests"
    git push origin main
    popd
    ${TOOLS}/flux reconcile source git flux-system -n flux-system
  fi
}

run_custom_scripts(){
    pwd
    echo "$(dirname "$0")"
  for f in "$(dirname "$0")"/custom/*.sh ; do
      echo "$f"
      if [ -x "$f" ] ; then
          echo executing "$f"
          $f
      fi
  done
}

main() {
    github_env_check
    do_capi
    do_flux
    create_local_values_file
    follow_capi_user_guide
    push_progressive_delivery_manifests_to_gitops_dev_repo
    run_custom_scripts
}

main

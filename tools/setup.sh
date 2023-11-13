#!/bin/bash

# This paves a cluster with the prerequisites for Weave GitOps
# Enterprise development:
# - a stock Flux installation, bootstrapped against the git repo given
#   in GITHUB_USER and GITHUB_REPO
# - a CAPI installation that includes the vcluster provider
#
# These are (mostly) idempotent -- calling this script repeatedly
# won't break anything (assuming your custom scripts are also safe to
# repeat).
#
# This script will get called by ./reboot.sh after recreating the kind
# cluster, and you can call it separately if you don't want to
# recreate the kind cluster first. This might be useful if ./reboot.sh
# repeatedly fails to make progress.

source $PWD/tools/flags.env

do_capi(){
  tool_check "clusterctl"

  EXP_CLUSTER_RESOURCE_SET=true ${TOOLS}/clusterctl init \
    --infrastructure vcluster
}

do_flux(){
  tool_check "flux"
  reset_clones_dir

  if [ "$DELETE_GITOPS_DEV_REPO" == "1" ]; then
    tool_check "gh"

    ${TOOLS}/gh repo delete "$GITHUB_USER/$GITHUB_REPO" --yes
  fi

  ${TOOLS}/flux bootstrap github \
    --owner="$GITHUB_USER" \
    --repository="$GITHUB_REPO" \
    --branch=main \
    --components-extra=image-reflector-controller,image-automation-controller \
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
  reset_clones_dir
  ${TOOLS}/gh repo clone "ssh://git@github.com/$GITHUB_USER/$GITHUB_REPO" "$CLONESDIR/$GITHUB_REPO"
  mkdir -p "$CLONESDIR/$GITHUB_REPO/clusters/bases/rbac"
  mkdir -p "$CLONESDIR/$GITHUB_REPO/clusters/bases/networkpolicy"
  cp "$(dirname "$0")/git-files/wego-admin.yaml" "$CLONESDIR/$GITHUB_REPO/clusters/bases/rbac/wego-admin.yaml"
  cp "$(dirname "$0")/git-files/flux-system-networkpolicy.yaml" "$CLONESDIR/$GITHUB_REPO/clusters/bases/networkpolicy/flux-system-networkpolicy.yaml"
  pushd "$CLONESDIR/$GITHUB_REPO"
  git add clusters/bases
  git commit -m "Add wego-admin role" || echo "No commit necessary; treating as a no-op"
  git push origin main
  popd
}

# Steps we ask you to do in https://docs.gitops.weave.works/docs/cluster-management/getting-started/
follow_capi_user_guide(){
    add_files_to_git
    kubectl delete --ignore-not-found secret my-pat
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
    reset_clones_dir
    ${TOOLS}/gh repo clone "ssh://git@github.com/$GITHUB_USER/$GITHUB_REPO" "$CLONESDIR/$GITHUB_REPO"
    mkdir -p "$CLONESDIR/$GITHUB_REPO/apps/progressive-delivery"
    cp -r "$(dirname "$0")/../../progressive-delivery/tools/extra-resources/" "$CLONESDIR/$GITHUB_REPO/apps/progressive-delivery/"
    rm "$CLONESDIR/$GITHUB_REPO/apps/progressive-delivery/istio/resources/gateway.yaml"
    cp "$(dirname "$0")/git-files/progressive-delivery-kustomizations.yaml" "$CLONESDIR/$GITHUB_REPO/clusters/management/progressive-delivery-kustomizations.yaml"
    pushd "$CLONESDIR/$GITHUB_REPO"
    git add apps/progressive-delivery
    git add clusters/management
    git commit -m "Add progressive-delivery manifests" || "No commit necessary; treating as a no-op"
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

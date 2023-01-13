#!/bin/bash

github_env_check() {
  if [[ -z "$GITHUB_TOKEN" ]]; then
    echo '!!! Missing GITHUB_TOKEN env var'
    exit 1
  fi
  if [[ -z "$GITHUB_REPO" ]]; then
    echo '!!! Missing GITHUB_REPO env var, e.g. my-github-user/dev-team'
    exit 1
  fi
  if [[ -z "$NAMESPACE" ]]; then
    echo '!!! Missing NAMESPACE env var'
    exit 1
  fi
}

bootstrap_source() {
  ssh-keygen -t ed25519 -C "$GITUB_REPO deploy key" -N "" -f /tmp/git-source-key

  gh repo deploy-key add -R "github.com/$GITHUB_REPO" /tmp/git-source-key.pub

  # replace / with - in GITHUB_REPO
  SOURCE_NAME="${GITHUB_REPO//\//-}"
  SECRET_NAME="$SOURCE_NAME-auth"

  flux create secret git $SECRET_NAME \
    --url=ssh://git@github.com/$GITHUB_REPO \
    --private-key-file=/tmp/git-source-key \
    --namespace $NAMESPACE

  flux create source git $SOURCE_NAME \
    --secret-ref $SECRET_NAME \
    --url ssh://git@github.com/$GITHUB_REPO \
    --branch main \
    --namespace $NAMESPACE
}

main() {
  github_env_check
  bootstrap_source
}

main
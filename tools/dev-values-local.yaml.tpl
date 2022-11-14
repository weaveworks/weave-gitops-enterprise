---
config:
  capi:
    repositoryURL: https://github.com/$GITHUB_USER/$GITHUB_REPO.git

extraEnvVars:
  - name: WEAVE_GITOPS_FEATURE_COST_ESTIMATION
    value: "${FEATURE_COST_ESTIMATION}"

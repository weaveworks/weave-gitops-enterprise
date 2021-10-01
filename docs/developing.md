# Developing `weave-gitops-enterprise`

[comment]: <> (Github can generate TOCs now see https://github.blog/changelog/2021-04-13-table-of-contents-support-in-markdown-files/)

A guide to making it easier to develop `weave-gitops-enterprise`. If you came here expecting but not finding an answer please make an issue to help improve these docs!

## Thing to explore in the future

- **tilt files** for faster feedback loops when interactively developing kubernetes services.
- ??

## How to change the code

### TDD

When making code modifications see if you can write a test first!

- **Integration and unit tests** should be places in the `_test.go` file next to the source you're modifying.
- **Acceptance tests** live in `./test/acceptance`

## How to run services locally against an existing cluster

Sometimes its nice to demo / experiment with the service(s) you're changing locally.

### The `capi-service`

_Note: the following instructions will use a new local database, you can probably reconcile the internal cluster database with the local one with some fancy fs mounting, tbd..._

The `capi-service` requires the presence of a valid entitlement secret for it to work. Make sure an entitlement secret has been added to the cluster and that the `capi-service` has been configured to look for it using the correct namespace/name. By default, entitlement secrets are named `weave-gitops-enterprise-credentials` and are added to the `wego-system` namespace. If that's not the case, you will need to point the service to the right place by explicitly specifying the relevant environment variables (example below).

An existing entitlement secret that you can use can be found [here](../test/utils/scripts/entitlement-secret.yaml). Alternatively, you can generate your own entitlement secret by using the `wge-credentials` binary.

```bash
# Optional, configure the kube context the capi-server should use
export KUBECONFIG=test-server-kubeconfig

# Run the server configured using lots of env vars
CAPI_CLUSTERS_NAMESPACE=default CAPI_TEMPLATES_NAMESPACE=default GIT_PROVIDER_TOKEN=$GITHUB_TOKEN GIT_PROVIDER_TYPE=github GIT_PROVIDER_HOSTNAME=github.com CAPI_TEMPLATES_REPOSITORY_URL=https://github.com/my-org/my-repo CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH=main ENTITLEMENT_SECRET_NAMESPACE=wego-system ENTITLEMENT_SECRET_NAME=weave-gitops-enterprise-credentials go run cmd/capi-service/main.go
```

You can query the local capi-server:

```bash
# via curl
curl http://localhost:8000/v1/credentials

# via the cli
go run cmd/mccp/main.go --endpoint http://localhost:8000/ templates list

# via the ui
cd ui-cra
CAPI_SERVER_HOST=http://localhost:8000 yarn start
```

### The `gitops-broker` service

```bash
go run cmd/gitops-repo-broker/main.go --db-uri /tmp/mccp.db --db-type sqlite --port 8090
```

You can query the local gitops-broker:

```bash
# via curl
curl http://localhost:8090/api/clusters

# via the ui
cd ui-cra
GITOPS_HOST=http://localhost:8090 yarn start
```

## How to update the version of `weave-gitops`

[`weave-gitops-enterprise`](https://github.com/weaveworks/weave-gitops-enterprise) depends on [`weave-gitops`](https://github.com/weaveworks/weave-gitops). When WG makes a new release we'll want to update the version WGE depends on. It goes a little something like this:

```bash
export WG_VERSION=0.2.4

# 1.update the backend golang code
go get github.com/weaveworks/weave-gitops@$WG_VERSION

# 2. Update the frontend typescript/javascript code
cd ui-cra && yarn add @weaveworks/weave-gitops@$WG_VERSION

# 3. Update the crds by copying all the files across to `./charts/mccp/templates/crds`
open "https://github.com/weaveworks/weave-gitops/tree/v${WG_VERSION}/manifests/crds"
```

## How to update the demo cluster

The demo cluster currently lives at http://34.67.250.163:30080/ but will hopefully move behind a DNS address with auth _soon_.

1. Figure out the version of chart you want to deploy. If we've done a release recently you can change it to `0.0.8` or a major version like that. To deploy a unreleased version from `main` or a `branch` we need to figure out the git ref version:

   1. Get your local copy of `weave-gitops-enterprise` up to date by running `git fetch`
   2. Figure out the git version ref of `origin/main` (for example) with: `git describe --always --match "v*" --abbrev=7 origin/main | sed 's/^[^0-9]*//'`. You could also provide `origin/fixes-the-funny-bug` as the branch name here.
   3. It will output a ref that looks like this: `0.0.7-10-g9838aff`

2. Update the deployed version on the test cluster

   1. As of writing the `HelmRelease` lives in [management/weave-gitops-enterprise/artifacts/mccp-chart/helm-chart/HelmRelease.yaml](https://github.com/wkp-example-org/capd-demo-simon/blob/main/management/weave-gitops-enterprise/artifacts/mccp-chart/helm-chart/HelmRelease.yaml), but may have moved, so look around for the helm-release file if this has gone missing.
   2. Find the `spec.chart.spec.version` field and change it to the desired value.
   3. If this is an official release (`0.0.9` etc) make sure the release repo is set:
      ```
       sourceRef:
         kind: HelmRepository
         name: weave-gitops-enterprise-mccp-chart-release
         namespace: wego-system
      ```
      Otherwise make sure its using dev:
      ```
       sourceRef:
         kind: HelmRepository
         name: weave-gitops-enterprise-mccp-chart-dev
         namespace: wego-system
      ```
   4. Commit to `main` or PR and merge to `main`.

3. Voila

## How to inspect/modify the `sqlite` database of a running cluster

Copy the database to your local machine and inspect using sqlite

```bash
kubectl cp mccp/mccp-cluster-service-79854d9fcb-bwvp7:/var/database/mccp.db mccp.db
sqlite mccp.db
```

Or, we can inspect _and modify_ the database in the cluster with

```bash
kubectl exec -ti -n mccp mccp-cluster-service-79854d9fcb-bwvp7 -- /bin/sh
apk add sqlite3
sqlite /var/database/mccp.db
```

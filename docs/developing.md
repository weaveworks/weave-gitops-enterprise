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

```bash
# Optional, configure the kube context the capi-server should use
export KUBECONFIG=test-server-kubeconfig

# Run the server configured using lots of env vars
CAPI_CLUSTERS_NAMESPACE=default CAPI_TEMPLATES_NAMESPACE=default GIT_PROVIDER_TOKEN=$GITHUB_TOKEN GIT_PROVIDER_TYPE=github GIT_PROVIDER_HOSTNAME=github.com CAPI_TEMPLATES_REPOSITORY_URL=https://github.com/my-org/my-repo CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH=main go run cmd/capi-service/main.go
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

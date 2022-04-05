# Developing `weave-gitops-enterprise`

A guide to making it easier to develop `weave-gitops-enterprise`. If you came here expecting but not finding an answer please make an issue to help improve these docs!

## Building the project

To build all binaries and containers use the following command:

```bash
# Builds everything
make

# Builds just the binaries
make binaries
```

If you encounter a build error for the containers which looks like this:

```log
 > [builder  7/15] RUN go mod download:
#15 20.58 go mod download: github.com/weaveworks/weave-gitops-enterprise-credentials@v0.0.1: invalid version: git ls-remote -q origin in /go/pkg/mod/cache/vcs/2d85ed3446e0807d78000711febc8f5eeb93fa1a010e290025afb84defca1ae6: exit status 128:
#15 20.58       remote: Invalid username or password.
#15 20.58       fatal: Authentication failed for 'https://github.com/weaveworks/weave-gitops-enterprise-credentials/'
------
executor failed running [/bin/sh -c go mod download]: exit code: 1
make: *** [cmd/event-writer/.uptodate] Error 1
```

Run make with the following postfix:

```bash
make GITHUB_BUILD_TOKEN=${GITHUB_TOKEN}
```

Further, don't forget to update your `~/.gitconfig` with:

```bash
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

## Thing to explore in the future

- **tilt files** for faster feedback loops when interactively developing kubernetes services.
- ??

## How to change the code

### TDD

When making code modifications see if you can write a test first!

- **Integration and unit tests** should be placed in the `_test.go` file next to the source you're modifying.
- **Acceptance tests** live in `./test/acceptance`

## How to run services locally against an existing cluster

Sometimes it's nice to demo / experiment with the service(s) you're changing locally.

### The `clusters-service`

_Note: the following instructions will use a new local database, you can probably reconcile the internal cluster database with the local one with some fancy fs mounting, tbd..._

To have entitlements, create a cluster and point your `kubectl` to it. It doesn't matter what kind of cluster you create.
Integration tests have a config located [here](../test/integration/test/kind-config.yaml) for inspiration.

The `clusters-service` requires the presence of a valid entitlement secret for it to work. Make sure an entitlement secret has been added to the cluster and that the `clusters-service` has been configured to look for it using the correct namespace/name. By default, entitlement secrets are named `weave-gitops-enterprise-credentials` and are added to the `wego-system` namespace. If that's not the case, you will need to point the service to the right place by explicitly specifying the relevant environment variables (example below).

An existing entitlement secret that you can use can be found [here](../test/utils/scripts/entitlement-secret.yaml). Alternatively, you can generate your own entitlement secret by using the `wge-credentials` binary.

#### Create a local database (optional):

```bash
$ (cd cmd/event-writer && go run main.go database create --db-type sqlite --db-uri file:///tmp/wge.db)
INFO[0000] created all database tables

# inspect db
$ sqlite3 /tmp/wge.db
SQLite version 3.28.0 2019-04-15 14:49:49
Enter ".help" for usage hints.

sqlite> .tables
alerts                 cluster_statuses       git_commits
capi_clusters          clusters               node_info
cluster_info           events                 pull_requests
cluster_pull_requests  flux_info              workspaces
sqlite>
```

#### Port forward the source-controller to access profiles (optional):

To query profiles the `clusters-service` needs to be able to DNS resolve the source-controller which provides the helm-repo (profile) info.

Goes like this: `/v1/profiles` on the `clusters-service` finds the `HelmRepository` CR to figure out the URL where it can get a copy of the `index.yaml` that lists all the profiles.

```yaml
kind: HelmRepository
status:
  # some url that only resolves when running inside the cluster
  url: source-controller.svc.wego-system/my-repo/index.yaml
```

Outside the cluster this is no good (e.g. `curl`ing the above URL will fail). To fix this we need to:

1. expose the source-controller outside the cluster (in another terminal):
   - `kubectl -n wego-system port-forward svc/source-controller 8080:80`
2. tell the `clusters-service` to forget about _most_ of that above URL it finds on the `HelmRepository` and use the port-forwarded one instead:
   - `SOURCE_CONTROLLER_LOCALHOST=localhost:8080`

#### Run the server:

```bash
# Optional, configure the kube context the capi-server should use
export KUBECONFIG=test-server-kubeconfig

# The weave-gitops core library uses an embedded Flux. That's not going to work when used as a library though
# so we need to tell it to use a different Flux. This is also done by the clusters-service deployment.
export WEAVE_GITOPS_FLUX_BIN_PATH=`which flux`

# If you have port-forward the source-controller from a cluster make sure to include its local address when starting the clusters-service:
SOURCE_CONTROLLER_LOCALHOST=localhost:8080

# Run the server configured using lots of env vars
DB_URI=/tmp/wge.db CAPI_CLUSTERS_NAMESPACE=default CAPI_TEMPLATES_NAMESPACE=default GIT_PROVIDER_TYPE=github GIT_PROVIDER_HOSTNAME=github.com CAPI_TEMPLATES_REPOSITORY_URL=https://github.com/my-org/my-repo CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH=main ENTITLEMENT_SECRET_NAMESPACE=wego-system ENTITLEMENT_SECRET_NAME=weave-gitops-enterprise-credentials go run cmd/clusters-service/main.go
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

## Developing the UI

We usually develop the UI against the test server and by default the UI dev server will use that.

```bash
cd ui-cra
yarn
yarn start
```

Open up http://localhost:3000. Changes to code will be hot-reloaded.

### UI against a local clusters-service

When you need to develop the UI against new features that haven't made to the test cluster yet you can run your own clusters-service locally and point the UI dev server at it with:

```bash
CAPI_SERVER_HOST=http://localhost:8000 yarn start
```

### UI against a remote server

If you need to use a remote server, set the `CAPI_SERVER_HOST` env var to the server's address before running `yarn start`:

```bash
CAPI_SERVER_HOST=http://34.67.250.163:30080 yarn start
```

### Testing changes to an unreleased weave-gitops locally

Maybe you need to add an extra export or tweak a style in a component in weave-gitops:

```bash
# build the weave-gitops ui-library
cd weave-gitops
git checkout cool-new-ui-feature
make ui-lib

# use it in wge
cd weave-gitops-enterprise/ui-cra

# optionally clean up node_modules a bit if changes don't seem to be coming through
rm -rf node_modules/@weaveworks/weave-gitops/

# install local copy of weave-gitops ui-lib
yarn add ../../weave-gitops/dist
```

## How to update the version of `weave-gitops`

[`weave-gitops-enterprise`](https://github.com/weaveworks/weave-gitops-enterprise) depends on [`weave-gitops`](https://github.com/weaveworks/weave-gitops). When WG makes a new release we'll want to update the version WGE depends on. It goes a little something like this:

```bash
export WG_VERSION=0.2.4

# 1.update the backend golang code
cd cmd/clusters-service
go get -d github.com/weaveworks/weave-gitops@$WG_VERSION
go mod tidy -compat=1.17
cd ../..
go mod tidy -compat=1.17

# 2. Update the frontend typescript/javascript code
cd ui-cra && yarn add @weaveworks/weave-gitops@$WG_VERSION

# 3. Update the crds by copying all the files across to `./charts/mccp/crds`
open "https://github.com/weaveworks/weave-gitops/tree/v${WG_VERSION}/manifests/crds"
```

## Demo clusters

We have 5 demo clusters currently that we use to demonstrate our work and test new features.

|                    UI                |                       GitOps                        |  CAPI  |
|--------------------------------------|-----------------------------------------------------|--------|
| http://34.67.250.163:30080           | https://github.com/wkp-example-org/capd-demo-simon  |  CAPD  |
| https://demo-01.wge.dev.weave.works  | https://gitlab.git.dev.weave.works/wge/demo-01      |  CAPG  |
| https://demo-02.wge.dev.weave.works  | https://github.com/wkp-example-org/demo-02          |   -    |
| https://demo-03.wge.dev.weave.works  | https://gitlab.git.dev.weave.works/wge/demo-03      |  CAPG  |
| https://demo-04.wge.dev.weave.works  | https://github.com/wkp-example-org/demo-04          |   -    |

---
**CAPI NAME COLLISION WARNING**

`demo-01` and `demo-03` are currently deployed on the same [GCP project](https://console.cloud.google.com/home/dashboard?project=wks-tests) so there may be collisions when creating CAPI clusters if they share the same name. Therefore avoid using common names like `test` and prefer to prefix them with your name i.e. `bob-test-2` instead.

---

There is no process to update these clusters automatically at the moment when there is a new release/merge to main although that would be desirable for a couple of them. The following sections describe how to get kubectl access to each of those clusters and how to update them to a newer version of Weave GitOps Enterprise.

#### 34.67.250.163

The test cluster currently lives at a static ip but will hopefully move behind a DNS address with auth _soon_.

Hit up http://34.67.250.163:30080

The private ssh key to the server lives in the `pesto test cluster ssh key` secret in 1Password.

1. Grab it and save it to `~/.ssh/cluster-key`
1. Set permissions `chmod 600 ~/.ssh/cluster-key`
1. Add it to your current ssh agent session with `ssh-add ~/.ssh/cluster-key`
1. Copy `kubeconfig` using this ssh key
   ```
   LANG=en_US.UTF-8 LC_ALL=en_US.UTF-8 scp wks@34.67.250.163:.kube/config demokubeconfig.txt
   ```
1. Port forward the api-server port (6443) in another tab
   ```
   ssh wks@34.67.250.163 -L 6443:localhost:6443
   ```
1. Use the `kubeconfig`:
   ```
   export KUBECONFIG=demokubeconfig.txt
   kubectl get pods -A
   ```

#### demo-01

Requires: gcloud CLI >= 352.0.0

Install and configure the gcloud CLI if needed. Then run:

```sh
gcloud container clusters get-credentials demo-01 --region europe-north1-a
```   

#### demo-02

Requires: gcloud CLI >= 352.0.0

Install and configure the gcloud CLI if needed. Then run:

```sh
gcloud container clusters get-credentials demo-02 --region europe-north1-a
```

#### demo-03

Requires: gcloud CLI >= 352.0.0

Install and configure the gcloud CLI if needed. Then run:

```sh
gcloud container clusters get-credentials demo-03 --region europe-north1-a
```

#### demo-04

Requires: aws CLI >= 2.5.2

Install and configure the aws CLI if needed. Then run:

```sh
aws eks --region eu-west-1 update-kubeconfig --name demo-04
```

### How to update to a new version

The following steps use [demo-01](https://demo-01.wge.dev.weave.works) as an example but the same concepts can be applied to all demo clusters. Depending on the cluster, you may need to sign up to our [on-prem Gitlab instance](https://gitlab.git.dev.weave.works) using your @weave.works email address and request access to the [Weave GitOps Enterprise](https://gitlab.git.dev.weave.works/wge) group or get added to the [wkp-example-org](https://github.com/wkp-example-org) in Github.

1. Figure out the version of the WGE chart you want to deploy:

   1. If we've done a release recently you can change it to `0.0.19` or a major version like that.
   2. Alternatively, to deploy an unreleased version from `main` or another branch you need to take a look at the [branch](#how-to-determine-the-version-of-a-branch) or the [charts repo](#how-to-search-for-a-helm-release-using-a-commit-sha) to determine the version.

2. Find the `HelmRelease` definition for WGE in the [repo](https://github.com/wkp-example-org/demo-01). It is called `weave-gitops-enterprise` and is part of the `wego-system` namespace. Locate the `spec.chart.spec.version` field [(example)](https://gitlab.git.dev.weave.works/wge/demo-01/-/blob/d431861309aae9c3645817af19c597c2f9d6f410/clusters/demo-01/wego-system/wego-system.yaml#L30) and update it to the new version (i.e. `0.0.17-88-ge4e540d`) by committing to `main` or via a PR.

   1. If this is an official release (i.e `0.0.19` etc) make sure the release repo is set:
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

3. Flux will detect this change and update the cluster with the version you specified in the previous step.

4. Voila

---
**NOTE FOR UPDATING 34.67.250.163**

As of writing the `HelmRelease` for 34.67.250.163 lives in [.weave-gitops/clusters/kind-kind/system/weave-gitops-enterprise.yaml](https://github.com/wkp-example-org/capd-demo-simon/blob/main/.weave-gitops/clusters/kind-kind/system/weave-gitops-enterprise.yaml), but may have moved, so look around for the Helm release file, if this has gone missing.

---

## How to determine the version of a branch

1. Get your local copy of `weave-gitops-enterprise` up to date by running `git fetch`

2. Figure out the git version ref of `origin/main` (for example) with: `git describe --always --match "v*" --abbrev=7 origin/main | sed 's/^[^0-9]*//'`. You could also provide `origin/fixes-the-funny-bug` as the branch name here.

3. It will output a ref that looks like this: `0.0.7-10-g9838aff`

## How to search for a Helm release using a commit sha

Requires: helm CLI >= 3.8.1

1. Add the charts repo locally:

```sh
helm repo add wkp https://charts.dev.wkp.weave.works/charts-v3 \
   --username wge --password gitops
```

2. Use the commit sha to find the relevant chart version by running the following:

```sh
helm repo update && helm search repo wkp --devel --versions | grep e4e540d
```
where `e4e540d` is your commit sha. This will return `wkp/mccp  	0.0.17-88-ge4e540d 	1.16.0     	A Helm chart for Kubernetes` where `0.0.17-88-ge4e540d` is the version you're looking for.


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

## How to make a self-signed cert that works in chrome!

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -sha256 -days 365 -nodes -subj '/CN=localhost' -addext "subjectAltName = DNS.1:localhost"
```

### MacOS trust it

```
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain cert.pem
```

### clusters service use it

```
clusters-service <OTHER_ARGS...> --tls-cert-file cert.pem --tls-private-key key.pem
```

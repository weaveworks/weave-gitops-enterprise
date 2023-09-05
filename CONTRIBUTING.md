# Developing `weave-gitops-enterprise`

A guide to making it easier to develop `weave-gitops-enterprise`. If you came here expecting but not finding an answer please make an issue to help improve these docs!

## The big picture

Weave GitOps Enterprise (WGE) is packaged as a Helm chart and currently consists
of the following components:

- `clusters-service`
  The API of WGE. This is the component that backend engineers will be changing
  most often. Uses the
  [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) to convert our
  gRPC service definitions to HTTP endpoints. Also imports API handlers from
  other projects (Weave GitOps among others) and exposes them to consumers.
- `ui-server`
  The UI of WGE. This is the component that frontend engineers will be changing
  most often. Built in React and uses yarn as the package manager.
- [cluster-bootstrap-controller](https://github.com/weaveworks/cluster-bootstrap-controller)
  Allows for custom Jobs to be executed on newly provisioned CAPI clusters. Most
  often, this will be used to install CNI which CAPI does not install. Without
  this controller, newly provisioned clusters would not be ready to be used by
  end users. Because it also references the CAPI CRD, it requires CAPI tooling
  to be installed first.
- [cluster-controller](https://github.com/weaveworks/cluster-controller)
  Defines the CRD for declaring leaf clusters. A leaf cluster is a cluster that
  the management cluster can query via a kubeconfig. This controller ensures
  that kubeconfig secrets have been supplied for leaf clusters. Because it also
  references the CAPI CRD, it requires CAPI tooling to be installed first.

## One-time setup

### If using Docker Desktop for Mac, scale up the machine

Docker Desktop is a complete Linux virtual machine, which will be used
for building and running the whole product - the defaults are way too
conservative.

Dedicate at least 16GB RAM and as many CPU cores as you can
spare - if you don't, everything will grind to a halt. Re-building the
whole backend takes about a minute or two when there's enough
resources, but if you're being too stingy with resources you could
easily find yourself waiting half an hour or more, every time you want
to change any code.

As for disk space, if you can spare a few hundred gigabytes, do give
it a few hundred gigabytes. Every time your development environment
loads a change, Docker will store another few megabytes of volumes which
adds up quicker than you'd think - and when you run out, you might get
annoying mystery errors.

After making this change, remember to turn your Docker Desktop off and
on again.

### Enable buildkit in docker (linux only)

You need to enable the buildkit feature in docker, or docker will complain:

> [BuildKit is enabled by default for all users on Docker Desktop](https://docs.docker.com/build/buildkit/#getting-started)

```bash
export DOCKER_BUILDKIT=1
```

### Github Personal Access Token

A PAT is needed to get access to our private repositories and packages.

This token needs at least the `repo` and `read:packages`
permissions. If you want to be able to delete the GitOps repo every
time you recreate your local Kind cluster, add the `delete_repo`
permission too and set the `DELETE_GITOPS_DEV_REPO` flag to 1. You
can create a token [here](https://github.com/settings/tokens), and
export it as:

```bash
export GITHUB_TOKEN=your_token
```

### Go package configuration

In order for Go to be able to download private dependencies, you must also
update your `~/.gitconfig` with:

```bash
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

If you are running into issues installing Go modules, try also setting the `GOPRIVATE` environment variable:

```bash
export GOPRIVATE=github.com/weaveworks/*
```

### Github user

You will also be using your personal GitHub account to host GitOps repositories. Therefore you need to export your GitHub username as well:

```bash
export GITHUB_USER=your_username
```

### Log in to GHCR

We use OCI artifacts hosted in GHCR, so you need your docker to be
logged in to this repository:

```bash
echo $GITHUB_TOKEN | docker login --username $GITHUB_USER --password-stdin ghcr.io
```

## Run a local development environment

To run a local development environment, you need to install
[Docker](https://www.docker.com) and
[kubectl](https://kubernetes.io/docs/reference/kubectl/kubectl/).

Run `make dependencies` to download binaries needed by the scripts in
`tools/`. These are saved in `tools/bin/`.

### Preparation

> :warning: The following script will **delete** a local Kind cluster named
> `wge-dev` and a remote repository named `wge-dev` in your personal GitHub
> account, if either of them exists. Take a look at the script to understand
> what it does and how to customize the cluster/repository names.

Run the following script to get a Kind cluster ready for Tilt:

```bash
./tools/reboot.sh
```

This will recreate a local Kind cluster, install CAPD and setup Flux to
reconcile from a GitOps repository in your personal GitHub account. It will also
create a file containing local settings such as your GitOps repository that the
enterprise Helm chart will use in the next step.

> If `reboot.sh` creates the Kind cluster but fails afterwards, you
> can use _just_ `./tools/setup.sh` to install CAPD and Flux and
> create the local config file. It is safe to run it repeatedly.

### Customizing your development environment

#### Custom kind configuration

The `reboot.sh` script has the capability to patch the default kind config
using custom configuration provided in the `./tools/custom/` directory. Place
any configuration you'd like in a file matching the pattern `kind-cluster-patch-*.yaml`
in that directory and it will get merged into the base configuration.

#### Custom scripts

The `reboot.sh` script will execute all scripts it finds in `./tools/custom/` that
match the file name pattern `*.sh` after creating the cluster and installing
all components.

#### Customizing Tilt

If you create a `Tiltfile.local` in the local directory, Tilt will detect and include any calls there, allowing you to add a bit of customization to your local environment. That's useful if you want to point to another external cluster, by default Tilt only allows local clusters like kind or use a different registry:

```
allow_k8s_contexts('another-cluster')

default_registry('my-custom-registry:5000')
```

### Start environment

To start the development environment, run

```bash
make cluster-dev
```

and your system should build and start. The first time you run this, it will
take ~10 mins (depending on your connection speed) to build all the containers
and deploy them to your local cluster. This is because the docker builds have to
download all the Go modules/JS libraries from scratch, use the Tilt UI to check
progress. Subsequent runs should be a lot faster.

When `cluster-service` has become green, you should be able to access
your cluster at [http://localhost:8000](http://localhost:8000). The login is
username `wego-admin` and password `dev`.

Any change you make to local code will trigger tilt to rebuild and restart the
pods running in your system.

### Faster frontend development

Especially for frontend development, the time it takes for the pod to restart
can be annoying. To spin up a local development frontend against your
development cluster, run:

```
yarn
PROXY_HOST=http://localhost:8000 yarn start
```

Now you have a separate frontend running on
[http://localhost:3000](http://localhost:3000) with in-process reload.

Tip: by starting tilt with the command `MANUAL_MODE=true tilt up` instead,
you will cause tilt to stop restarting the backend unless you click
the button in the UI. If you find that the backend always restarts
while you're doing something, this might help.

### When something goes wrong

Here's a list of things that might go wrong.

#### The tilt "Uncategorized" tab is red, and says something about "addons.cluster.x-k8s.io"

This means you have not installed a CAPI provider. If you set up your
cluster with `tools/reset.sh` then it should have been installed
already.

If you want to fix it without a cluster reset, look for the function
running `clusterctl` in `tools/reset.sh`.

If that doesn't help, it could also mean that your version of
`clusterctl` is too old, so you simply have the wrong version of the
cluster API. Note that `make dependencies` does not support upgrades,
so running it does not keep you up to date - you have to manually
delete `tools/bin/clusterctl` and then re-run `make dependencies` to
upgrade it.

#### Errors mentioning "Unauthorized" when building docker images

Note: this only applies to the building step - not after it's started!

This might mean your github token isn't set correctly. Try turning
tilt off, running `export GITHUB_TOKEN=your_token` in that terminal,
and starting it again and see if that fixes it.

If that doesn't work, make sure your token has _both_ repo _and_
package permissions, and that it hasn't expired.

If you still experience problems, ask someone in your team if it works
for them, as it could be a new dependency you don't have permissions
to read. If it's stopped working for you both, you probably need to
change the github permissions.

#### The UI is empty and you see an error mentioning "kubernetes client initialization failed: Unauthorized"

Whenever the `Uncategorized` job runs in tilt, it resets all
permissions. When that happens, it can delete the permissions for the
`chart-mccp-cluster-service` pod that's currently running.

Manually re-run the `Uncategorized` job to make sure it runs as
expected. When it's finished, re-start the
`chart-mccp-cluster-service` deployment. Once that's finished, it
should work again.

#### Tilt keeps printing "too many open files"

If you're on a mac, it should go away if you simply turn Docker
Desktop off and on again - you don't have to delete the machine, just
reboot. If you're on linux, try turning the kind docker
container off and on again, or reboot your machine.

To get it to go away, on mac try to put `ulimit -n <big_number>` in
your shell config file (probably `~/.zshrc` or maybe `~/.bashrc`) -
half a million or so is a big number.

On linux, you might need to set `fs.nr_open` to a big value in
/etc/sysctl.conf, and you might need to override LimitNOFILE for the
containerd systemd unit. There's other related settings that can cause
similar-looking errors - the `fs.inotify` family of settings often
start up very low and might need a bump.

## Building the project

To build all containers use the following command:

```bash
# Builds everything - make sure you exported GITHUB_TOKEN as shown in
# the one-time setup
make GITHUB_BUILD_TOKEN=${GITHUB_TOKEN}
```

## Common dev workflows

The following sections suggest some common dev workflows.

### Tooling

Before you start working on the code, you need to install the following tools:

- [Go](https://go.dev/dl/) (1.18) for backend development
- [Node.js](https://nodejs.org/en/download/releases/) (14) for frontend development
- [kubectl](https://kubernetes.io/docs/tasks/tools/) for interacting with Kubernetes clusters
- [Helm](https://helm.sh/docs/intro/install/) for working with Helm charts
- [Buf](https://docs.buf.build/installation) for generating code from protobuf definitions

### How to do local dev on the API

Most of the code for the API is under `./cmd/clusters-service`. There's a
Makefile in that directory with some helpful targets so when working on the API
make sure to run these from that location instead of root. The following
commands assume execution from `./cmd/clusters-service`.

To install gRPC tooling run:

```bash
make install
```

The API endpoints are defined as gRPC definitions in
`./cmd/clusters-service/api/capi_server.proto`. Therefore if you need to add or
update an endpoint you need to first define it in there. For example the
following endpoint is used to return the version of WGE.

```proto
// GetEnterpriseVersion returns the WeGO Enterprise version
rpc GetEnterpriseVersion(GetEnterpriseVersionRequest)
  returns (GetEnterpriseVersionResponse){
    option (google.api.http) = {
      get: "/v1/enterprise/version"
  };
}
```

After making a change in the protobuf definition, you will need to run `make generate` to regenerate the code.

To run the service locally, run:

```bash
export CAPI_CLUSTERS_NAMESPACE=default
go run main.go

```

You can execute HTTP requests to the API by pointing to an endpoint, for
example:

```bash
curl --insecure http://localhost:8000/v1/enterprise/version
```

The --insecure flag is needed because the service will generate self-signed
untrusted certificates by default.

To run all unit tests before pushing run:

```bash
make unit-tests
```

To run all tests, including integration tests run:

```bash
make test
```

### Creating leaf cluster

To create leaf clusters to test out our features, we can rely on the [vcluster](https://www.vcluster.com/) to help us deploy new clusters on the fly. That project will basically create a entire cluster inside you kind cluster without adding much overhead.

to get started install the `vcluster` cli first, by following https://www.vcluster.com/docs/getting-started/setup and then just run the `./tools/create-leaf-cluster.sh` script.

```shell
$ ./tools/create-leaf-cluster.sh leaf-cluster-01
```

This command will create a new cluster and configure the `GitopsCluster` CR pointing to the cluster's kubeconfig.

Note that this won't configure completelly the cluster, you might need to install flux and rbac rules in order to be able to query it properly. But it should be already visible on the Weave Gitops cluster's tab.

### How to install everything from your working branch on a cluster

When you push your changes to a remote branch (even before creating a PR for
it), CI will kick off a build that runs a quick suite of tests and then builds
your containers and creates a new Helm chart tagged with the most recent commit.
This Helm chart includes all the changes from your branch and can be used to
deploy WGE as a whole to a cluster.

1. Find the version of the Helm chart you need to deploy:

   ```bash
   # Add the Helm repo locally (needs to happen only once)
   ./tools/bin/helm repo add weave-gitops-enterprise-charts \
      https://charts.dev.wkp.weave.works/charts-v3 \
      --username wge --password gitops
   "weave-gitops-enterprise-charts" has been added to your repositories_

   # Search the Helm repo for the commit SHA that corresponds to your most recent commit
   ./tools/bin/helm repo update > /dev/null 2>&1 \
      && ./tools/bin/helm search repo weave-gitops-enterprise-charts --devel --versions \
      | grep <commit-SHA>
   weave-gitops-enterprise-charts/mccp      <chart-version-with-commit-SHA>     1.16.0          A Helm chart for Kubernetes
   ```

2. Create a new kind cluster and install flux

   ```bash
   cat > kind-cluster-with-extramounts.yaml <<EOF
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   nodes:
   - role: control-plane
     extraMounts:
     - hostPath: /var/run/docker.sock
       containerPath: /var/run/docker.sock
   EOF

   ./tools/bin/kind create cluster \
       --name kind \
       --config=kind-cluster-with-extramounts.yaml
   export GITHUB_TOKEN=<your-GH-token>
   ./tools/bin/flux bootstrap github \
       --owner=<your-GH-username> \
       --repository=config \
       --personal=true \
       --path=clusters/kind
   ```

3. Install CAPI

   ```bash
   ./tools/bin/clusterctl init --infrastructure docker
   ```

4. Install WGE

   ```bash
   cat > values.yaml <<EOF
   tls:
     enabled: true
   config:
     capi:
       repositoryURL: <your config repo URL>
   EOF

   kubectl apply -f ./test/utils/data/entitlement/entitlement-secret.yaml
   ./tools/bin/flux create source helm weave-gitops-enterprise-charts \
       --url=https://charts.dev.wkp.weave.works/charts-v3 \
       --namespace=flux-system \
       --secret-ref=weave-gitops-enterprise-credentials
   ./tools/bin/flux create hr weave-gitops-enterprise \
       --namespace=flux-system \
       --interval=10m \
       --source=HelmRepository/weave-gitops-enterprise-charts \
       --chart=mccp \
       --chart-version=<chart-version-with-commit-SHA> \
       --values values.yaml
   ```

## How to change the code

### TDD

When making code modifications see if you can write a test first!

- **Integration and unit tests** should be placed in the `_test.go` file next
  to the source you're modifying.
- **Acceptance tests** live in `./test/acceptance`

#### Grab a copy of the SQLite chart cache

The `clusters-service` caches the `index.yaml` and `values.yaml` files from
helm-repos (profiles) in a SQLite database. This is to avoid hammering the
source-controller with requests. The cache is stored in a file called
`/tmp/helm-cache/charts.db` by default.

```
kubectl cp flux-system/$(kubectl get pods -A -l app=clusters-service --no-headers -o custom-columns=":metadata.name"):/tmp/helm-cache/charts.db mccp.db
```

## Developing the UI

We usually develop the UI against the test server and by default the UI dev
server will use that.

```bash
yarn
yarn start
```

Open up http://localhost:3000. Changes to code will be hot-reloaded.

### Unit Tests

To start the `jest` test runner CLI dialog:

```
$ yarn test

PASS  src/components/Applications/__tests__/index.test.tsx
  Applications index test
    ✓ renders table rows (349 ms)
    snapshots
      ✓ loading (170 ms)
      ✓ success (168 ms)

Test Suites: 1 passed, 1 total
Tests:       3 passed, 3 total
Snapshots:   2 passed, 2 total
Time:        5.448 s
Ran all test suites.

Watch Usage: Press w to show more.
```

#### UI Unit Test Tips

- The `@testing-library/react` package provides a test renderer as well as helpers for dealing with hooks and component state
- Snapshots alone generally aren't enough, you should do some assertions to validate component behavior
- Hooks can be tested in isolation from components using the `act` helper.

#### Snapshot Tests

We use a technique called "Snapshots" to record the rendered output of components and track them in version control over time. Snapshots are not really tests, since they don't have any explicity assertions. Think of them more as a record of the output of a component.

When combined with the `styled-components` integration, snapshots give us a way to track styling logic over time. This can be very helpful in debugging styling issues that would otherwise be hard to understand.

After any changes to styling logic, you should expect to update snapshots, else unit tests will fail.

To update snapshots:

```
yarn test -u
```

### How to do local dev on the UI

The easiest way to dev on the UI is to use an existing cluster.
[demo-01](https://demo-01.wge.dev.weave.works/) is kept automatically up to date
with every change that lands on main. To use it run the following command:

```bash
PROXY_HOST=https://demo-01.wge.dev.weave.works/ yarn start
```

The username/password used to login are stored in
[1Password](https://start.1password.com/open/i?a=ALD7KP6DEJGYREYHXRNYI3F7KY&v=xdzphlycic6bzwggrot2y73jaa&i=jz6ytgxay7ktq5vg6w2wlu3m2i&h=weaveworks.1password.com).

If you need to develop the UI against new features that haven't made to the test
cluster yet, you can run your own clusters-service locally and point the UI dev
server at it with:

```bash
PROXY_HOST=http://localhost:8000 yarn start
```

**Recommended UI Development environment variables**:

```bash
# This will point your local FE dev server to the local cluster backend
export WEGO_EE_PROXY_HOST=http://localhost:8000
# This will skip building the UI docker container, which we don't need because we have a dev server.
export SKIP_UI_BUILD=true
# This will use your local Go compiler to build the binary, rather than building in a docker container via Tilt.
# You may need to install things like gcc to make this work, since we are using CGO=true.
export NATIVE_BUILD=true
```

### Testing changes to an unreleased weave-gitops locally

It is possible to test against local files from gitops OSS during
development. If you want to do this, you need to make sure you've
started tilt using `MANUAL_MODE=true tilt up` so that it won't
re-start your backend. After that's running, you have to do
development against the `yarn start` method.

For another method that has fewer caveats, look at "How to update
`weave-gitops` to a non-released version during development".

One magical command to "reload" core (assumes the project directories are located in the same directory):

```bash
# Note, this assumes you have core and EE at the same level in the file system
make core-ui && make core-lib
```

## How to update `weave-gitops` to a released version

[`weave-gitops-enterprise`](https://github.com/weaveworks/weave-gitops-enterprise) depends on [`weave-gitops`](https://github.com/weaveworks/weave-gitops). When WG makes a new release we'll want to update the version WGE depends on. It goes a little something like this:

```bash
export WG_VERSION=0.2.4

# 1.update the backend golang code
go get -d github.com/weaveworks/weave-gitops@$WG_VERSION
go mod tidy

# 2. Update the frontend typescript/javascript code
yarn add @weaveworks/weave-gitops@$WG_VERSION
```

## How to update `weave-gitops` to a non-released version during development

This will update WGE to use the latest `main` of `weave-gitops`

```bash
make update-weave-gitops
```

You can also pick a different branch as long as that branch has a PR
associated with it by running

```bash
make update-weave-gitops BRANCH=<my-branch-name>
```

## How to update the version of `cluster-controller`

When a new release of the cluster-controller is made we'll usually want to update it in WGE.

```bash
export CC_VERSION=1.2.0

# update the backend golang code in the weave-gitops-enterprise repo root
cd weave-gitops-enterprise
go get -d github.com/weaveworks/cluster-controller@CC_VERSION
go mod tidy
```

Copy across the newer helm-chart

```bash
cd ../cluster-controller

# generates and copies a new helm subchart to ../weave-gitops-enterprise
make helm
cd ../weave-gitops-enterprise

# TODO: improve this:
# Carefully add back the "important" changes to the chart, read the comments as you go.
git add --patch
```

## How to update the version of `cluster-bootstrap-controller`

Update `images.clusterBootstrapController` in https://github.com/weaveworks/weave-gitops-enterprise/blob/main/charts/mccp/values.yaml

Manually copy across any big changes to the deployment or CRDs from cluster-controller/config into weave-gitops-enterprise/charts/mccp/

## Demo clusters

We have 3 demo clusters currently that we use to demonstrate our work and test
new features.

| UI                                  | GitOps                                                | CAPI |
| ----------------------------------- | ----------------------------------------------------- | ---- |
| http://35.188.40.143:30080          | https://github.com/wkp-example-org/capd-demo-reloaded | CAPD |
| https://demo-01.wge.dev.weave.works | https://gitlab.git.dev.weave.works/wge/demo-01        | CAPG |

---

## Managing multiple clusters

As enterprise features are deployed, the multi-cluster permissions may need to be updated as well. For example viewing canaries from a leaf cluster did not work. Below is an example rbac config that resolved the canary issue:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: demo-02
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: impersonate-user-groups
subjects:
  - kind: ServiceAccount
    name: demo-02
    namespace: default
roleRef:
  kind: ClusterRole
  name: user-groups-impersonator
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: user-groups-impersonator
rules:
  - apiGroups: [""]
    resources: ["users", "groups"]
    verbs: ["impersonate"]
  - apiGroups: [""]
    resources: ["namespaces"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apiextensions.k8s.io"] # required for canary support
    resources: ["customresourcedefinitions"]
    verbs: ["get", "list", "watch"]
```

#### demo-01

Requires: gcloud CLI >= 352.0.0

Install and configure the gcloud CLI if needed. Then run:

```bash
gcloud container clusters get-credentials demo-01 --region europe-north1-a
```

### How to update to a new version

The following steps use [demo-03](https://demo-03.wge.dev.weave.works) as an
example but the same concepts can be applied to all demo clusters. Depending on
the cluster, you may need to sign up to our [on-prem Gitlab
instance](https://gitlab.git.dev.weave.works) using your @weave.works email
address and request access to the [Weave GitOps
Enterprise](https://gitlab.git.dev.weave.works/wge) group or get added to the
[wkp-example-org](https://github.com/wkp-example-org) in Github.

1. Figure out the version of the WGE chart you want to deploy:

   1. If we've done a release recently you can change it to `0.0.19` or a major
      version like that.
   2. Alternatively, to deploy an unreleased version from `main` or another
      branch you need to take a look at the
      [branch](#how-to-determine-the-version-of-a-branch) or the [charts
      repo](#how-to-search-for-a-helm-release-using-a-commit-sha) to determine
      the version.

2. Find the `HelmRelease` definition for WGE in the [repo](https://github.com/wkp-example-org/demo-03).
   It is called `weave-gitops-enterprise` and is part of the `flux-system`
   namespace. Locate the `spec.chart.spec.version` field
   [(example)](https://gitlab.git.dev.weave.works/wge/demo-03/-/blob/77390541343d889f0fab0fc50198f6f233692003/clusters/demo-03/wego-system/wego-system.yaml#L30)
   and update it to the new version (i.e. `0.0.17-110-g485f9bf`) by committing
   to `main` or via a PR.

   1. If this is an official release (i.e `0.0.19` etc) make sure the release repo is set:
      ```
       sourceRef:
         kind: HelmRepository
         name: weave-gitops-enterprise-mccp-chart-release
         namespace: flux-system
      ```
      Otherwise make sure its using dev:
      ```
       sourceRef:
         kind: HelmRepository
         name: weave-gitops-enterprise-mccp-chart-dev
         namespace: flux-system
      ```

3. Flux will detect this change and update the cluster with the version you
   specified in the previous step.

4. Voila

---

> **Note**
>
> As of writing the `HelmRelease` for 35.188.40.143 lives in
> https://github.com/wkp-example-org/capd-demo-reloaded/blob/main/clusters/management/weave-gitops-enterprise.yaml
> but may have moved, so look around for the Helm release file, if this has gone missing.

---

## How to determine the version of a branch

1. Get your local copy of `weave-gitops-enterprise` up to date by running `git fetch`

2. Figure out the git version ref of `origin/main` (for example) with:
   `git describe --always --match "v*" --abbrev=7 origin/main | sed 's/^[^0-9]*//'`.
   You could also provide `origin/fixes-the-funny-bug` as the branch name here.

3. It will output a ref that looks like this: `0.0.7-10-g9838aff`

## How to search for a Helm release using a commit sha

Requires: helm CLI >= 3.8.1

1. Add the charts repo locally:

```bash
./tools/bin/helm repo add wkp https://charts.dev.wkp.weave.works/charts-v3 \
   --username wge --password gitops
```

2. Use the commit sha to find the relevant chart version by running the following:

```bash
./tools/bin/helm repo update \
  && ./tools/bin/helm search repo wkp --devel --versions \
  | grep e4e540d
```

where `e4e540d` is your commit sha. This will return `wkp/mccp 0.0.17-88-ge4e540d 1.16.0 A Helm chart for Kubernetes` where `0.0.17-88-ge4e540d` is the version you're looking for.

## How to search for a Helm release from GCP OCI registry

1. If you are using a Helm verion prior to `v3.8.0` set the `HELM_EXPERIMENTAL_OCI` environment variable. Helm versions `v3.8.0` and newer have OCI support enabled by default

```bash
export HELM_EXPERIMENTAL_OCI=1
```

2. If you haven't already, install and configure the [gcloud CLI](https://cloud.google.com/sdk/docs/install)

3. Use the gcloud cli to query registry artifacts
   > The Google Artifact Registry Docker repository can hold both helm charts and docker images. If both types will be deployed to the same registry, charts should be stored in the `charts` namespace and images in the `images` namespace as documented [here](https://cloud.google.com/artifact-registry/docs/helm)

```bash
gcloud artifacts docker images list europe-west1-docker.pkg.dev/weave-gitops-clusters/weave-gitops-enterprise --include-tags
```

4. Once you know the version tag you can use the oci image url and version to run helm show/pull/install commands
   > With oci registries the `--version` flag is required

```bash
helm show all oci://europe-west1-docker.pkg.dev/weave-gitops-clusters/weave-gitops-enterprise/charts/mccp --version 0.8.1-55
```

## How to make a self-signed cert that works in chrome!

```bash
openssl req -x509 \
    -newkey rsa:4096 \
    -keyout key.pem -out cert.pem \
    -sha256 -days 365 \
    -nodes -subj '/CN=localhost' \
    -addext "subjectAltName = DNS.1:localhost"
```

### MacOS trust it

```bash
sudo security add-trusted-cert \
    -d -r trustRoot \
    -k /Library/Keychains/System.keychain cert.pem
```

### clusters service use it

```bash
clusters-service <OTHER_ARGS...> --tls-cert-file cert.pem --tls-private-key key.pem
```

## How to get a kubeconfig for an AKS cluster

Requires: azure CLI >= 2.36.0

Install and configure the azure CLI if needed. Then run:

```bash
az aks get-credentials --name <cluster-name> --resource-group <resource-group> --admin
```

## How to get a kubeconfig for an EKS cluster

Requires: aws CLI >= 2.5.2

Install and configure the aws CLI if needed. Then run:

```bash
aws eks --region <aws-region> update-kubeconfig --name <cluster-name>
```

## How to add a feature flag

First add a new feature flag in [values.yaml](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/charts/mccp/values.yaml) ([example](https://github.com/weaveworks/weave-gitops-enterprise/blob/0605d2992fed7888dd2c5045d675c7baeadb03ef/charts/mccp/values.yaml#L28))

For example:

```yaml
# Turns on pipelines features if set to true. This includes the UI elements.
enablePipelines: false
```

Then add an `if` block to the [deployment.yaml](https://github.com/weaveworks/weave-gitops-enterprise/blob/main/charts/mccp/templates/clusters-service/deployment.yaml) ([example](https://github.com/weaveworks/weave-gitops-enterprise/blob/27e143749b5cda32d6d8ac6eadab3d1e2db2999e/charts/mccp/templates/clusters-service/deployment.yaml#L61-L64)) template of cluster-service in order to conditionally set an environment variable. This variable will be available to cluster-service at runtime. As a convention, the environment variable needs to be prefixed with `WEAVE_GITOPS_FEATURE_` and takes a value of `true`.

For example:

```yaml
{{- if .Values.enablePipelines }}
- name: WEAVE_GITOPS_FEATURE_PIPELINES
  value: "true"
{{- end }}
```

Finally, you can use the feature flag from Go through the `featureflags` package, for example:

```go
if featureflags.Get("WEAVE_GITOPS_FEATURE_PIPELINES") != "" {
  // Feature flag has been set
}
```

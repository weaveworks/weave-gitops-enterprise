# Developing `weave-gitops-enterprise`

A guide to making it easier to develop `weave-gitops-enterprise`. If you came here expecting but not finding an answer please make an issue to help improve these docs!

## Tooling
You will need the following tools installed on your local workstation:
- Go 1.17 
  
  `brew install go@1.17`
- Kubectl 1.23
  
  `brew install kubectl@1.23`
- Helm 3.8.2
  
  `brew install helm`
- Kind 0.12.0 for creating a local Kubernetes cluster
  
  `brew install kind`
- Flux 0.29.5 for setting up Flux on a cluster
  
  `brew install fluxcd/tap/flux`
- Buf 1.4.0 for generating code from protobuf definitions
  
  `brew install bufbuild/buf/buf`
- clusterctl 1.1.3 for installing CAPI components on the cluster
  
  `brew install clusterctl`

## The big picture
Weave GitOps Enterprise (WGE) currently consists of the following components:
- clusters-service
  The API of WGE. This is the component that backend engineers will be changing most often. Uses the [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway) to convert our gRPC service definitions to HTTP endpoints.
- ui-server
  The UI of WGE. This is the component that frontend engineers will be changing most often. Built in React and uses yarn as the package manager.
- [cluster-bootstrap-controller](https://github.com/weaveworks/cluster-bootstrap-controller)
  Allows for custom Jobs to be executed on newly provisioned CAPI clusters. Most often, this will be used to install CNI which CAPI does not install. Without this controller, newly provisioned clusters would not be ready to be used by end users. Because it also references the CAPI CRD, it requires CAPI tooling to be installed first.
- [cluster-controller](https://github.com/weaveworks/cluster-controller)
  Defines the CRD for declaring leaf clusters. A leaf cluster is a cluster that the management cluster can query via a kubeconfig. This controller ensures that kubeconfig secrets have been supplied for leaf clusters. Because it also references the CAPI CRD, it requires CAPI tooling to be installed first.
- event-writer **Soon to be deprecated**
  
  Subscribes to a NATS queue and listens for messages sent by the wkp-agent. These messages carry cluster information about leaf clusters which is then written to a file-based SQLite db.
- wkp-agent **Soon to be deprecated**
  
  Agent installed by end users to management cluster and leaf clusters with the purpose of reporting back cluster information to management cluster.

As part of the chart, other dependencies also get installed:
- NATS **Soon to be deprecated**
  
  Agents running on the management cluster or on leaf clusters send events to a NATS queue. event-writer receives these events from the queue and does its thing.

## One-time setup
You need a github Personal Access Token to build the service. This
token needs at least the `repo` and `read:packages`
permissions. You can create one
[here](https://github.com/settings/tokens), and export it as:
```bash
export GITHUB_TOKEN=your_token
```

You must also update your `~/.gitconfig` with:
```bash
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```

You will also be using your personal GitHub account to host GitOps repositories. Therefore it is useful to export your GitHub username as well:
```bash
export GITHUB_USER=your_username
```

Finally, make sure you can access https://github.com/weaveworks/weave-gitops-enterprise-credentials

## Building the project

To build all containers use the following command:

```bash
# Builds everything - make sure you exported GITHUB_TOKEN as shown in
# the one-time setup
make GITHUB_BUILD_TOKEN=${GITHUB_TOKEN}
```

## Common dev workflows
The following sections suggest some common dev workflows.

### How to do local dev on the API

Most of the code for the API is under `./cmd/clusters-service`. There's a Makefile in that directory with some helpful targets so when working on the API make sure to run these from that location instead of root. The following commands assume execution from `./cmd/clusters-service`.

To install gRPC tooling run:
```sh
make install
```

The API endpoints are defined as gRPC definitions in `./cmd/clusters-service/api/capi_server.proto`. Therefore if you need to add or update an endpoint you need to first define it in there. For example the following endpoint is used to return the version of WGE.

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

To run the service locally,  run:
```sh
export CAPI_CLUSTERS_NAMESPACE=default
go run main.go
```

You can execute HTTP requests to the API by pointing to an endpoint, for example:
```sh
curl --insecure https://localhost:8000/v1/enterprise/version
```
The --insecure flag is needed because the service will generate self-signed untrusted certificates by default.

To run all unit tests before pushing run:
```sh
make unit-tests
```

To run all tests, including integration tests run:
```sh
make test
```

### How to do local dev on the controllers
When working on controllers it's often easier to run them locally against kind clusters to avoid causing issues on shared clusters (i.e. demo-01) that may be used by other engineers. To install the CRDs on your local kind cluster run:

```sh
make install
```

To run the controller locally run:
```sh
make run
```

To run all tests before pushing run:
```sh
make test
```

### How to install everything from your working branch on a cluster
When you push your changes to a remote branch (even before creating a PR for it), CI will kick off a build that runs a quick suite of tests and then builds your containers and creates a new Helm chart tagged with the most recent commit. This Helm chart includes all the changes from your branch and can be used to deploy WGE as a whole to a cluster.

1. Find the version of the Helm chart you need to deploy:
   ```sh
   # Add the Helm repo locally (needs to happen only once)
   helm repo add weave-gitops-enterprise-charts https://charts.dev.wkp.weave.works/charts-v3 --username wge --password gitops
   _"weave-gitops-enterprise-charts" has been added to your repositories_
   # Search the Helm repo for the commit SHA that corresponds to your most recent commit
   helm repo update > /dev/null 2>&1 && helm search repo weave-gitops-enterprise-charts --devel --versions | grep <commit-SHA>
   weave-gitops-enterprise-charts/mccp      <chart-version-with-commit-SHA>     1.16.0          A Helm chart for Kubernetes
   ```
   
2.  Create a new kind cluster and install flux
    ```sh
    cat > kind-cluster-with-extramounts.yaml <<EOF
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
      extraMounts:
      - hostPath: /var/run/docker.sock
        containerPath: /var/run/docker.sock
    EOF
    
    kind create cluster --name kind --config=kind-cluster-with-extramounts.yaml
    export GITHUB_TOKEN=<your-GH-token>
    flux bootstrap github --owner=<your-GH-username> --repository=config --personal=true --path=clusters/kind
    ```

3.  Install CAPI
    ```sh
    clusterctl init --infrastructure docker
    ```

4.  Install WGE
    ```sh
    cat > values.yaml <<EOF
    tls:
      enabled: true
    config:
      capi:
        repositoryURL: <your config repo URL>
    EOF
   
    kubectl apply -f ./test/utils/scripts/entitlement-secret.yaml
    flux create source helm weave-gitops-enterprise-charts --url=https://charts.dev.wkp.weave.works/charts-v3 --namespace=flux-system --secret-ref=weave-gitops-enterprise-credentials
    flux create hr weave-gitops-enterprise --namespace=flux-system --interval=10m --source=HelmRepository/weave-gitops-enterprise-charts --chart=mccp --chart-version=<chart-version-with-commit-SHA> --values values.yaml

## Run a local development environment
To run a local development environment, you need to install the following tools:
 * [docker](https://www.docker.com)
 * [kind](https://kind.sigs.k8s.io/)
 * [tilt](https://tilt.dev/)
 * [flux](https://fluxcd.io/docs/installation/)
 * [clusterctl](https://cluster-api.sigs.k8s.io/user/quick-start.html#install-clusterctl)

### Preparation
Run the following command to get a Kind cluster ready for Tilt:
```sh
./tools/reboot.sh
```

This will recreate a local Kind cluster, install CAPD and setup Flux to reconcile from a GitOps repository in your personal GitHub account. It will also create a file containing local settings such as your GitOps repository that the enterprise Helm chart will use in the next step.

### Start environment
To start the development environment, run
```sh
tilt up
```
and your system should build and start. The first time you run this, it will take ~20-30 mins (depending on your connection speed) to build all the containers and deploy them to your local cluster. This is because the docker builds have to download all the Go modules/JS libraries from scratch, use the Tilt UI to check progress.

When `chart-mccp-cluster-service` has become green, you should be able
to access your cluster at
[https://localhost:8000](https://localhost:8000). The login is
username `wego-dev` and password `dev`.

Any change you make to local code will trigger tilt to rebuild and
restart the pods running in your system.

**THINGS TO WATCH OUT FOR**

- If a change in your local settings results in a ConfigMap update, you will need to restart the `clusters-service` pod in order for the pod to read the updated ConfigMap.
- Every time you restart `clusters-service` it will generate new self-signed certificates, therefore you will need to reload the UI and accept the new certificate. Check for TLS certificate errors in the `chart-mccp-cluster-service` logs and if necessary re-trigger an update to rebuild it.

### Faster frontend development
Especially for frontend development, the time it takes for the pod to
restart can be annoying. To spin up a local development frontend against
your development cluster, run:
```
cd ui-cra
yarn
PROXY_HOST=https://localhost:8000 yarn start
```

Now you have a separate frontend running on
[http://localhost:3000](http://localhost:3000) with in-process reload.

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

The `clusters-service` requires the presence of a valid entitlement secret for it to work. Make sure an entitlement secret has been added to the cluster and that the `clusters-service` has been configured to look for it using the correct namespace/name. By default, entitlement secrets are named `weave-gitops-enterprise-credentials` and are added to the `flux-system` namespace. If that's not the case, you will need to point the service to the right place by explicitly specifying the relevant environment variables (example below).

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
  url: source-controller.svc.flux-system/my-repo/index.yaml
```

Outside the cluster this is no good (e.g. `curl`ing the above URL will fail). To fix this we need to:

1. expose the source-controller outside the cluster (in another terminal):
   - `kubectl -n flux-system port-forward svc/source-controller 8080:80`
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
DB_URI=/tmp/wge.db CAPI_CLUSTERS_NAMESPACE=default CAPI_TEMPLATES_NAMESPACE=default GIT_PROVIDER_TYPE=github GIT_PROVIDER_HOSTNAME=github.com CAPI_TEMPLATES_REPOSITORY_URL=https://github.com/my-org/my-repo CAPI_TEMPLATES_REPOSITORY_BASE_BRANCH=main ENTITLEMENT_SECRET_NAMESPACE=flux-system ENTITLEMENT_SECRET_NAME=weave-gitops-enterprise-credentials go run cmd/clusters-service/main.go
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

### How to do local dev on the UI

The easiest way to dev on the UI is to use an existing cluster. [demo-01](https://demo-01.wge.dev.weave.works/) is kept automatically up to date with every change that lands on main. To use it run the following command:

```sh
PROXY_HOST=https://demo-01.wge.dev.weave.works/ yarn start
```

The username/password used to login are stored in [1Password](https://start.1password.com/open/i?a=ALD7KP6DEJGYREYHXRNYI3F7KY&v=xdzphlycic6bzwggrot2y73jaa&i=jz6ytgxay7ktq5vg6w2wlu3m2i&h=weaveworks.1password.com).

If you need to develop the UI against new features that haven't made to the test cluster yet, you can run your own clusters-service locally and point the UI dev server at it with:

```sh
PROXY_HOST=http://localhost:8000 yarn start
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

We have 3 demo clusters currently that we use to demonstrate our work and test new features.

|                    UI                |                       GitOps                        |  CAPI  |
|--------------------------------------|-----------------------------------------------------|--------|
| http://34.67.250.163:30080           | https://github.com/wkp-example-org/capd-demo-simon  |  CAPD  |
| https://demo-01.wge.dev.weave.works  | https://gitlab.git.dev.weave.works/wge/demo-01      |  CAPG  |
| https://demo-02.wge.dev.weave.works  | https://github.com/wkp-example-org/demo-02          |  CAPG  |

---
**CAPI NAME COLLISION WARNING**

`demo-01` and `demo-02` are currently deployed on the same [GCP project](https://console.cloud.google.com/home/dashboard?project=wks-tests) so there may be collisions when creating CAPI clusters if they share the same name. Therefore avoid using common names like `test` and prefer to prefix them with your name i.e. `bob-test-2` instead.

---

`demo-01` is automatically updated to the latest version of main. `demo-02` is manually updated to the latest release of Weave GitOps Enterprise. The following sections describe how to get kubectl access to each of those clusters and how to update them to a newer version of Weave GitOps Enterprise.

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

### How to update to a new version

The following steps use [demo-03](https://demo-03.wge.dev.weave.works) as an example but the same concepts can be applied to all demo clusters. Depending on the cluster, you may need to sign up to our [on-prem Gitlab instance](https://gitlab.git.dev.weave.works) using your @weave.works email address and request access to the [Weave GitOps Enterprise](https://gitlab.git.dev.weave.works/wge) group or get added to the [wkp-example-org](https://github.com/wkp-example-org) in Github.

1. Figure out the version of the WGE chart you want to deploy:

   1. If we've done a release recently you can change it to `0.0.19` or a major version like that.
   2. Alternatively, to deploy an unreleased version from `main` or another branch you need to take a look at the [branch](#how-to-determine-the-version-of-a-branch) or the [charts repo](#how-to-search-for-a-helm-release-using-a-commit-sha) to determine the version.

2. Find the `HelmRelease` definition for WGE in the [repo](https://github.com/wkp-example-org/demo-03). It is called `weave-gitops-enterprise` and is part of the `flux-system` namespace. Locate the `spec.chart.spec.version` field [(example)](https://gitlab.git.dev.weave.works/wge/demo-03/-/blob/77390541343d889f0fab0fc50198f6f233692003/clusters/demo-03/wego-system/wego-system.yaml#L30) and update it to the new version (i.e. `0.0.17-110-g485f9bf`) by committing to `main` or via a PR.

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
where `e4e540d` is your commit sha. This will return `wkp/mccp      0.0.17-88-ge4e540d      1.16.0          A Helm chart for Kubernetes` where `0.0.17-88-ge4e540d` is the version you're looking for.


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

## How to get a kubeconfig for an AKS cluster

Requires: azure CLI >= 2.36.0

Install and configure the azure CLI if needed. Then run:

```sh
az aks get-credentials --name <cluster-name> --resource-group <resource-group> --admin
```

## How to get a kubeconfig for an EKS cluster

Requires: aws CLI >= 2.5.2

Install and configure the aws CLI if needed. Then run:

```sh
aws eks --region <aws-region> update-kubeconfig --name <cluster-name>
```

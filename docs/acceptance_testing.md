
# Running acceptance tests locally

This is a guide to quickly setup the environment to run and debug acceptance locally running on a kind cluster. The test host can be either macOS or linux e.g. Ubuntu machine.

There are some prerequisites before running acceptance tests locally. It includes installing required tools and environment configurations. One can install and configure them as per their existing environment.

  ## Tools  & Utilities

It is recommended to install latest and stable version of these tools. All tools must be on path.
| Tool | Purpose | Installation
|--|--|--|
| Docker | Containers runtime environment | `https://docs.docker.com/get-docker` |
| Kind | Running local Kubernetes cluster | `https://kind.sigs.k8s.io/docs/user/quick-start#installation` |
|Kubectl|Kubernetes command-line tool| `https://kubernetes.io/docs/tasks/tools/install-kubectl-linux` |
| Helm | Package manager for Kubernetes | `https://helm.sh/docs/intro/install` |
| Clusterctl | Cluster API management command-line tool | `curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.0.3/clusterctl-darwin-amd64 -o clusterctl` <br> `curl -L https://github.com/kubernetes-sigs/cluster-api/releases/download/v1.0.3/clusterctl-ubuntu-amd64 -o clusterctl` <br> `chmod +x ./clusterctl` <br> `sudo mv ./clusterctl /usr/local/bin/clusterctl` |
| Totp-cli | Generates OTP tokens for two factor authentication | `wget https://github.com/yitsushi/totp-cli/releases/download/v1.1.17/totp-cli-v1.1.17-darwin-amd64.tar.gz` <br> `wget https://github.com/yitsushi/totp-cli/releases/download/v1.1.17/totp-cli-v1.1.17-ubuntu-amd64.tar.gz` <br> `tar -xf totp-cli-v1.1.17-darwin-amd64.tar.gz` <br> `mv ./totp-cli /usr/local/bin` |
| Selenium server | Standalone server for web browser instance | `wget https://selenium-release.storage.googleapis.com/3.14/selenium-server-standalone-3.14.0.jar` <br> `It is not required if test host is a macOS machine.`|
| flux | Command-line interface to bootstrap and interact with Flux | `https://fluxcd.io/docs/installation/#install-the-flux-cli`|

## Environment Setup

<font size="5">**Git**</font>

Configure git with the following global settings. It will elevates the manual intervention of certain git operations.

    git config --global init.defaultBranch main  
    git config --global user.email <your email address>  
    git config --global user.name <your user name>  
    git config --global url.git@github.com/.insteadOf https://github.com/  
    git config --global url.git@gitlab.com:.insteadOf https://gitlab.com/  
    git config --global url.git@gitlab.git.dev.weave.works:.insteadOf https://gitlab.git.dev.weave.works/ 


<font size="5">**Git provider(s) key fingerprints**</font>

Add git providers i.e. (GitHub, gitlab and gitlab-on-prm) key fingerprints to the known_hosts file.  

```
# Clean up potentially old keys
ssh-keygen -R github.com
ssh-keygen -R gitlab.com
ssh-keygen -R gitlab.git.dev.weave.works

# Append fresh new keys
ssh-keyscan gitlab.git.dev.weave.works github.com gitlab.com >> ~/.ssh/known_hosts
```

<font size="5">**Environment variables**</font>

These environment variables are needed by the setup scripts and tests to identify the run time test specifications e.g. cluster type, git provider etc.
```
export SELENIUM_DEBUG=true
export GITOPS_BIN_PATH=/usr/local/bin/gitops
export ARTIFACTS_BASE_DIR=/tmp/acceptance-tests
export MANAGEMENT_CLUSTER_KIND=kind
export CAPI_PROVIDER=capd
export EXP_CLUSTER_RESOURCE_SET=true
export UI_NODEPORT=30080
export MANAGEMENT_CLUSTER_CNAME=weave.gitops.enterprise.com
export CLUSTER_REPOSITORY=gitops-testing
export LOGIN_USER_TYPE=cluster-user
```

You can either set 'LOGIN_USER_TYPE' to `oidc` if oidc user authentication is desired or `cluster-user` if cluster user i.e. `wego-admin` account authentication is desired to run the tests.
You can also set 'CAPI_PROVIDER' to 'capa' or 'capg' as per requirements.

**User login**
```
export CLUSTER_ADMIN_PASSWORD=wego-enterprise
export CLUSTER_ADMIN_PASSWORD_HASH='$2b$12$1mxK92n85K.Zga8eNe2j.eqhtC2HnSrvbOk9obSVKbUgJm4SCMwFm'
export OIDC_ISSUER_URL='https://dex-01.wge.dev.weave.works'
export DEX_CLIENT_ID='weave-gitops-enterprise'
export DEX_CLIENT_SECRET='2JPIcb5IvO1isJ3Zii7jvjqbUtLtTC'
```

OIDC provider instance `https://dex-01.wge.dev.weave.works`  is already setup and ready to use for development and testing purposes.

You can set the environment variables for any one of the gitprovider as per your testing requirements.

**Github**
```
export GIT_PROVIDER=github
export GIT_PROVIDER_HOSTNAME=github.com
export GITHUB_ORG=<github org name>
export GITHUB_TOKEN=<github account token>
export GITHUB_USER=<github account user name>
export GITHUB_PASSWORD=<github account password>
export TOTP_TOKEN=<github MFA token key>
```
You must setup`MFA` for GitHub and export the MFA key as `TOTP_TOKEN`. It is required for automated GitHub authentication flows.

**Gitlab saas**
```
export GIT_PROVIDER=gitlab
export GIT_PROVIDER_HOSTNAME=gitlab.com
export GITHUB_ORG=<gitlab group name>
export GITHUB_TOKEN=<gitlab account token>
export GITHUB_USER=<gitlab account user name>
export GITHUB_PASSWORD=<gitlab account password>
export GITLAB_CLIENT_ID=<gitlab oath app id>
export GITLAB_CLIENT_SECRET=<gitlab oath app secret>
```
  
**Gitlab on-prem**
```
export GIT_PROVIDER=gitlab
export GIT_PROVIDER_HOSTNAME=gitlab.git.dev.weave.works
export GITHUB_ORG=<gitlab group name>
export GITHUB_TOKEN=<gitlab account token>
export GITHUB_USER=<gitlab account user name>
export GITHUB_PASSWORD=<gitlab account password>
export GITLAB_CLIENT_ID=<gitlab oath app id>
export GITLAB_CLIENT_SECRET=<gitlab oath app secret>
export GITOPS_GIT_HOST_TYPES="gitlab.git.dev.weave.works=gitlab"
export GITLAB_HOSTNAME=“gitlab.git.dev.weave.works"
```
You can use any gitlab on-prem instance to run tests. However, `gitlab.git.dev.weave.works` instance is already setup and ready to use for development and testing purposes.
You must configure the gitlab oath application with redirect url as below. It is required for automated gitlab authentication flows (applicabel to both gilab saas and gitlab on-prem).
    http://weave.gitops.enterprise.com:30080/oauth/gitlab

`weave.gitops.enterprise.com` is set as `MANAGEMENT_CLUSTER_CNAME` environment variable. Redirect url domain should match `MANAGEMENT_CLUSTER_CNAME` and `UI_NODEPORT`.

## Running Tests

- Run selenium server if test host is a linux machine. Selenium server is not required for macOS.

	`java -jar ./selenium-server-standalone-3.14.0.jar &`
- ***Command shell:*** Change directory to weave-gitops-enterprise. All paths in the following instructions are relative to `weave-gitops-enterprise` directory.

	`cd $HOME/go/src/github.com/weaveworks/weave-gitops-enterprise`	 
- Delete any existing kind cluster(s).

	`kind delete clusters --all`
- Create a new clean kind cluster.

	`kind create cluster  --config test/utils/data/local-kind-config.yaml` 

- ***Automatic installation:*** Test frame work automatically installs the  core and enterprise controllers and setup the management cluster along with required repository, resources, secrets and entitlements etc. Any subsequent test runs will skip the management cluster setup and starts the test execution straight away. You need to recreate the kind cluster in case you want to install new enterprise version/release for testing.

You may needed to add a `MANAGEMENT_CLUSTER_CNAME` entry to `/etc/hosts` file e.g. `192.168.0.5 weave.gitops.enterprise.com` (where `192.168.0.5` is test host's ip address) before start running the tests.

- ***Manual installation:*** You can manually install and setup core and enterprise controllers without running acceptance test. You must create the config repository i.e. `CLUSTER_REPOSITORY` prior to running the following command. The core controllers can not be installed if `CLUSTER_REPOSITORY` doesn't exists. Manual creation of cluster repository is only required for manual installation. 

	You may be prompted for administrator password while running the below script. It is needed to add a `MANAGEMENT_CLUSTER_CNAME` entry to `/etc/hosts` file e.g. `192.168.0.5 weave.gitops.enterprise.com` (where `192.168.0.5` is test host's ip address).

	`test/utils/scripts/wego-enterprise.sh setup $(pwd)`
	

- ***Enterprise chart version:*** The management cluster setup script tries to fetch the helm chart from *S3* corresponding  to latest commit hash of the working branch. In case if the image with latest commit hash doesn’t exist in *S3*, then you can manually override the chart version of your choice by setting `ENTERPRISE_CHART_VERSION` environment variable.  

	`export ENTERPRISE_CHART_VERSION=0.0.17-53-gb6aa363`

	If you make any changes to UI or backend, you need to rebuild the cluster. The easiest and fastest way is to push to origin (your remote branch). It will build the image corresponding to your local branch commit hash and push it to *S3*.
	You can also manually build and push the build to *S3*.

## Troubleshooting

Please refer to the Cluster API troubleshooting guide for issues related to `capd`. You may encounter following issues:
- Failed clusterctl init - ‘failed to get cert-manager object' (Resolved in latest docker version)

	`https://cluster-api.sigs.k8s.io/user/troubleshooting.html#failed-clusterctl-init---failed-to-get-cert-manager-object`
- Cluster API with Docker Desktop - “too many open files”  (Resolved in latest docker version)

	`https://cluster-api.sigs.k8s.io/user/troubleshooting.html#cluster-api-with-docker-desktop---too-many-open-files`

## Command Line run

You can run all or selected set of acceptance tests from command line.

**Examples:**

- It will only run tests with ‘@git’ label in the their description

	`go test -ginkgo.focus=@git -ginkgo.v —timeout=99999s`
- It will run all tests excluding  those which have ‘@gce|@eks|@application’ label in the their description.

	`go test -ginkgo.skip=@gce|@eks|@application -ginkgo.v —timeout=99999s`
- It will run all tests without focusing or skipping any tests. However if any test has been focused in the test code, then only those test(s) will be run.

	`go test -ginkgo.v —timeout=99999s`

In order to focus or skip tests from execution directly, just put `F` or `X` respectively in front of the `It` clause.

- Only this test will be run when test suite is run without `-ginkgo.focus` or `ginkgo.skip` flags.

	`FIt("@integration @git Verify pull request can be created for capi template to the management cluster", func() {`
- This test will be skipped when test suite is run without -ginkgo.focus or ginkgo.skip flags.

	`XIt("@integration @git Verify pull request can be created for capi template to the management cluster", func() {`

## VS Code run

Running test from VS Code is useful as it enables you to debug the tests.

**VS Code setup**

- Run -> Add configuration ->
- Use the below launch settings to configure VS Code. Paste and adjust these settings accordingly.
	```
	{
		// Use IntelliSense to learn about possible attributes.
		// Hover to view descriptions of existing attributes.
		// For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
		"version": "0.2.0",
		"configurations": [
			{
				"name": "Launch test function",
				"type": "go",
				"request": "launch",
				"mode": "test",
				"program": "${workspaceFolder}/test/acceptance/test/",
				"args": [
				"-ginkgo.v",
				],
				"env": {
				"GITLAB_TOKEN": "xxxxxxxxx",
				"GITOPS_BIN_PATH": "/usr/local/bin/gitops",
				"DOCKER_IO_USER": "xxxx",
				"DOCKER_IO_PASSWORD": "xxxx",
				"GIT_PROVIDER": "github",
				"GITHUB_ORG": "xxxxx",
				"CLUSTER_REPOSITORY": "gitops-testing",
				}
			}
		]
	}
	```
- The `arg` section above exactly behaves like the `go test -ginkgo.v` command line parameter(s).
- The `env` section above is not mandatory, it is to override any existing or missing environment variables in the shell.
- Make sure `MANAGEMENT_CLUSTER_CNAME` entry exists in `/etc/hosts` file. Since VS Code is a UI application, it has no sudo privileges and can not edit the hosts file.

	Example: `192.168.0.5 weave.gitops.enterprise.com` (where `192.168.0.5` is test host ip address)
-   Add breakpoints where you want the test to stop
-   Run -> Start debugging
-   View -> Debug console (for viewing test output)
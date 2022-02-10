
# Running acceptance locally

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
| Gitops | Gitops command line interface | `wget https://weave-gitops.s3.us-east-2.amazonaws.com/gitops-macOS-latest` <br> `wget https://weave-gitops.s3.us-east-2.amazonaws.com/gitops-ubuntu-latest` <br> `mv gitops-ubuntu-latest /usr/local/bin/gitops` <br> `sudo chmod +x /usr/local/bin/gitops`|

**Clusterctl workaround**

Run the following script to override downloading of cert-manager.yaml during capd infrastructure provider installation.

    cat > $HOME/.cluster-api/clusterctl.yaml <<- EOM
    cert-manager:
    	url: "https://github.com/cert-manager/cert-manager/releases/latest/cert-manager.yaml"
    EOM

## Environment Setup

<font size="5">**Git**</font>

Configure git with the following global settings. It will elevates the manual intervention of certain git operations.

    git config --global init.defaultBranch main  
    git config --global user.email <your email address>  
    git config --global user.name <your user name>  
    git config --global url.ssh://git@github.com/.insteadOf https://github.com/  
    git config --global url.git@gitlab.com:.insteadOf https://gitlab.com/  
    git config --global url.git@gitlab.git.dev.wkp.weave.works:.insteadOf https://gitlab.git.dev.wkp.weave.works/ 


<font size="5">**Git provider(s) key fingerprints**</font>

Add git providers i.e. (GitHub, gitlab and gitlab-on-prm) key fingerprints to the known_hosts file.  

```
cat > ~/.ssh/known_hosts <<- EOM
# github.com:22 SSH-2.0-babeld-a73e1397
github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
# github.com:22 SSH-2.0-babeld-a73e1397
github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=
# github.com:22 SSH-2.0-babeld-a73e1397
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
# gitlab.com:22 SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2
gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9
# gitlab.com:22 SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2
gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=
# gitlab.com:22 SSH-2.0-OpenSSH_7.9p1 Debian-10+deb10u2
gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf
gitlab.git.dev.wkp.weave.works ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBeP3Ucl8fV9dQK3cuN8A8z1t5ah9Xwq/oVGG/MOyiE4DyCP/m+XIBJ07dW7rl5cCpGivmzURslQuSQWLM7oSMVddzKMeMA+A2Xqf+c5jEpCDx08TarBInInqjgO3Yt9NmptQ0JsQgNYLugclQVcuk832/2Ge7M9kw8Dp9SeYsIG/8oBl8DeSXp7AR21zsnH0uKRil7a8I6Nmo8wC3s8iAj1KP/dYTn0S7M+8ZYM0ubrUyKULqAWMAH2KXG4fs2Z3yaK4yWugCre8KTSF2YJYsnNkfNy2NHyb/nGgIDLP3Or0ER5mRPqUgu1vgvXIk0nVKmYfGnWvByg5e2sn4QjE/
gitlab.git.dev.wkp.weave.works ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBB7Mqp5balgSUzdzgveNGarJatOw6elpMtKzawdtY+ugxWFNLskxoEydYqZFHDaS8D/bH1XvZYUemRZBd7vntbk=
gitlab.git.dev.wkp.weave.works ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPUuEL+yd8UwVnrP2PLSNslqfOt7o6WGDdEcD2f8SwRB
EOM
```

<font size="5">**Environment variables**</font>

These environment variables are needed by the setup scripts and tests to identify the run time test specifications e.g. cluster type, git provider etc.
```
export SELENIUM_DEBUG=true
export GITOPS_BIN_PATH=/usr/local/bin/gitops
export ARTIFACTS_BASE_DIR=/tmp/acceptance-tests
export MANAGEMENT_CLUSTER_KIND=kind
export ACCEPTANCE_TESTS_DATABASE_TYPE=sqlite
export EXP_CLUSTER_RESOURCE_SET=true
export UI_NODEPORT=30080
export NATS_NODEPORT=31490
export MANAGEMENT_CLUSTER_CNAME: weave.gitops.enterprise.com
export CLUSTER_REPOSITORY=gitops-testing
```

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

**Gitlab**
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
export GIT_PROVIDER_HOSTNAME=gitlab.git.dev.wkp.weave.works
export GITHUB_ORG=<gitlab group name>
export GITHUB_TOKEN=<gitlab account token>
export GITHUB_USER=<gitlab account user name>
export GITHUB_PASSWORD=<gitlab account password>
export export GITLAB_CLIENT_ID=<gitlab oath app id>
export GITLAB_CLIENT_SECRET=<gitlab oath app secret>
export GITOPS_GIT_HOST_TYPES="gitlab.git.dev.wkp.weave.works=gitlab"
export GITLAB_HOSTNAME=“gitlab.git.dev.wkp.weave.works"
```
You can use any gitlab on-prem instance to run tests. However, `gitlab.git.dev.wkp.weave.works` instance is already setup and ready to use for development and testing purposes.
You must configure the gitlab oath application with redirect url as below. It is required for automated gitlab authentication flows (applicabel to both gilab sas and gitlab on-prem).
    http://weave.gitops.enterprise.com:30080/oauth/gitlab

`weave.gitops.enterprise.com` is set as `MANAGEMENT_CLUSTER_CNAME` environment variable. Redirect url domain should match `MANAGEMENT_CLUSTER_CNAME`.


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

- ***Manual installation:*** You can manually install and setup core and enterprise controllers without running acceptance test. You must create the config repository i.e. `CLUSTER_REPOSITORY` prior to running the following command. The core controllers can not be installed if `CLUSTER_REPOSITORY` doesn't exists. Manuall creation of cluster repository is only required for manual installation. 

	You may be be prompted for administrator password while running the below script. It is needed to add a `MANAGEMENT_CLUSTER_CNAME` entry to `/etc/hosts` file e.g. `192.168.0.5 weave.gitops.enterprise.com` (where `192.168.0.5` is test host's ip address).

	`test/utils/scripts/wego-enterprise.sh setup $(pwd)`
	

- ***Enterprise chart version:*** The management cluster setup script tries to fetch the helm chart from *S3* corresponding  to latest commit hash of the working branch. In case if the image with latest commit hash doesn’t exist in *S3*, then you can manually override the chart version of your choice by setting `ENTERPRISE_CHART_VERSION` environment variable.  

	`export ENTERPRISE_CHART_VERSION=0.0.17-53-gb6aa363`

	If you make any changes to UI or backend, you need to rebuild the cluster. The easiest and fastest way is to push to origin (your remote branch). It will build the image corresponding to your local branch commit hash and push it to *S3*.
	You can also manually build and push the release to *S3*.



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
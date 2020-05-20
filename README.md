# Weave Kubernetes Subscription (WKS)

![Service description diagram](https://www.weave.works/assets/images/blt1670b4d9d8010619/KB_support_diagram.jpg)

## Purpose

This repository is to keep together work done on the Weaveworks Kubernetes Subscription. Track the progress in the [Github project](https://github.com/weaveworks/wks/projects/1).

## Important documents

See:

- [Meeting Notes](https://drive.google.com/open?id=1wfN4V6T9t1-eapXGabFZqkBCxyKW3uVZzz-cBCosgxs)
- [Phase 1 Plan](https://docs.google.com/document/d/1q3y0jDrzNKpTxPUi5JYf8vaPDTLV9_Ur65lxZFElDSo/edit)
  - [Pharos/WKS analysis](https://docs.google.com/document/d/1FRJd5Uj0CuHPwHbqXooIpUF1UKTy9tjsBaNqAA5BtrQ/edit)
  - [Test plan](https://docs.google.com/spreadsheets/d/1EdSdbdbFrYrjLwr33qAMF31n_g2hrSgogljen8RBHj4/edit)
- [WKS manifest draft](https://docs.google.com/document/d/1WtIE11RC-6f4mhp2Krsf1AsNCNEHcSuEQNp12nV0mDU/edit#)
- [Press release](https://www.weave.works/press/releases/weaveworks-launches-enterprise-gitops-services/)
- [Product Page](https://www.weave.works/product/enterprise-kubernetes-support/)
- [Theory of Gitops](https://docs.google.com/document/d/1Y8kr3gROHUnFuGR3h4adjwWH6E3ttGHIYwVuWWVv2VE/edit)
- [WKS Future](https://docs.google.com/document/d/1HK6r5CA0ZlUQT3PmFWVQ_93TlPz31nHdx13-pve1S4U/edit#)

## Kerberos

- [Explain like I'm 5: Kerberos](http://www.roguelynn.com/words/explain-like-im-5-kerberos/)
- [About Kerberos Principals and Keys](https://ssimo.org/blog/id_016.html)
- [MIT Kerberos Documentation](http://web.mit.edu/kerberos/krb5-1.12/doc/index.html)

## Notes

### Releasing

To release a new version of the project:
- Verify that we have an updated dependencies file
  - run `GITHUB_TOKEN=<your token> bin/sca-generate-deps.sh`
  - If there are changes to the file `user-guide/content/deps/_index.md` merge this **before** creating the release
- Create a new tag: `git tag -a 1.0.1` The -a is important as it creates an annotated tag which is then used as a version number for builds.
- Push tag: `git push origin 1.0.1`
- CI will push binary to weaveworks-wkp.s3.amazonaws.com/wk-1.0.1
- Edit release notes https://github.com/weaveworks/wks/releases/edit/1.0.1
- Update rpm/wk.spec version and changelog
- Build an rpm `cd rpm && ./build wk.spec`
- Sign rpm: `rpm --addsign output/x86_64/wk-1.1.0-0.x86_64.rpm`
- Publish rpm to our yum repo https://github.com/weaveworks/rpm
  - Copy rpm in `wks/rhel/7`
  - `cd wks/rhel/7 && createrepo .`

### `tools/`

The `tools` directory is copied via `git subtree` from the
[build-tools](https://github.com/weaveworks/build-tools) repo.

### code-generator

```console
make gen
```

## Development

### Dependencies

- [`hugo`](https://gohugo.io/getting-started/quick-start/)
- [`embedmd`](https://github.com/campoy/embedmd)
- [`yarn`](https://classic.yarnpkg.com/en/docs/install)

### Build

```console
make
```

#### Upgrading the build image

- Update `build/Dockerfile` as required.
- Test the build locally:

```console
rm wks-build/.uptodate
make !$
```

- Push this change, get it reviewed, and merge it to `master`.
- Run:

```console
$ git checkout master ; git fetch origin master ; git merge --ff-only master
$ rm build/.uptodate
$ make !$
[...]
Successfully built deadbeefcafe
Successfully tagged docker.io/weaveworks/wks-build:latest
docker tag docker.io/weaveworks/wks-build docker.io/weaveworks/wks-build:master-XXXXXXX
touch build/.uptodate
$ docker push docker.io/weaveworks/wks-build:$(tools/image-tag)
```

- Update `.circleci/config.yml` to use the newly pushed image.
- Push this change, get it reviewed, and merge it to `master`.

### Documentation

Run:

```console
$ ./cmd/wk/wk user-guide --entitlements ./entitlements/2018-08-31-weaveworks.entitlements
INFO[0000] User guide server now running. Please open the following address in your browser: http://localhost:8080
```

Go to: [http://localhost:8080](http://localhost:8080)

# Using with a config repo instead of cluster and machine yaml files

We will create a cluster by pulling the cluster and machine yaml from git. We perform all the master node setup of today.

The following are new commandline arguments to `wk apply` which will result in a cluster being created.

- **git-url** The git repo url containing the cluster and machine yaml
- **git-branch** The branch within the repo to pull the cluster info from
- **git-deploy-key** The deploy key configured for the write access to the git repo

The new commandline arguments will be passed instead of --cluster and --machines.

```console
$ wk apply
  --git-url git@github.com:meseeks/config-repo.git \
  --git-branch dev \
  --git-deloy-key-path ./deploy-key
```

Using the url, branch, and deploy key, we will clone the repo - if we can't clone the repo we will error out.

These `--git` arguments are then used to setup and configure [flux](https://www.weave.works/oss/flux/) to automate cluster management.

We will rely on the user installing [fluxctl](https://docs.fluxcd.io/en/latest/references/fluxctl/) to interact with flux directly instead of trying to replicate the functionality within `wk`

# Running an end-to-end test against EKS using Quickstart

To run a test that uses the `wk-quickstart-eks` repo to construct a cluster and ensure that the correct pods are running, go to the `test/integration/test` directory and type: `go test --timeout=99999s`. The test will run for 20-25 minutes and then delete the cluster and any repositories (local and remote) created during the test.

The following environment variables must be set before running the test:

- GIT_PROVIDER_USER (username used to determine where to create an empty remote git repo)
- GIT_DEPLOY_KEY (private key that will be used to access the git repo)
- DOCKER_IO_USER (for fetching images)
- DOCKER_IO_PASSWORD (for fetching images)
- WKP_CLUSTER_COMPONENTS_IMAGE (used to manage components)

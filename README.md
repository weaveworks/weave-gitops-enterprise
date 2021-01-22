[![Coverage Status](https://coveralls.io/repos/github/weaveworks/wks/badge.svg?branch=master&t=Yv0GnU)](https://coveralls.io/github/weaveworks/wks?branch=master)

# Weave Kubernetes Platform

<img src="/docs/images/wk-transparent.svg" height="100">

Weave Kubernetes Platform (WKP) creates production ready clusters with ease and
facilitates GitOps: all configuration is in files which can be kept under version control.

## Getting Started

See our documentation by opening the user guide:

```console
wk user-guide
INFO[0000] User guide server now running. Please open the following address in your browser: http://localhost:8080
```

Also see:

- the `./docs` directory of this repository
- the [WKP Product Information](https://www.notion.so/weaveworks/WKP-Product-information-a6f142ce885b41c288ab97b0eb21fbf4) in notion

## Tracks

Depending on your requirements and how much infrastructure you already have in place you can choose a WKP `track`:

- `wks-ssh` - you have some existing machines and would like to install a Kubernetes cluster and cluster components onto them.
- `wks-components` - you have an existing Kubernetes cluster and would like to install the cluster components.
- `eks` - WKP will create an EKS cluster and install the cluster components.
- `wks-footloose` - [footloose](https://github.com/weaveworks/footloose) will be used to create virtual machines locally and then install a Kubernetes cluster and cluster components.

Diagram of the components in the `ssh` and `footloose` track:

![Component breakdown diagram](/docs/images/component-breakdown.png)

For the EKS version, see [EKS component breakdown](/docs/component-diagram-eks.md)

## Entitlements and Dockerhub credentials

Entitlements files are necessary to run `wk`. They contain the limits
granted by the commercial agreement with Weaveworks.

- By default `wk` will look in the artefacts directory `~/.wks/entitlements`.
- Alternatively, set `WKP_ENTITLEMENTS` to the entitlements file path:

```console
export WKP_ENTITLEMENTS=/path/to/file.entitlements
```

For development purposes valid entitlements files can be found in the `entitlements` directory of this repo.

In order to pull the WKP component images, the docker user specified in `setup/config.yaml` needs
to be granted access to the WKP dockerhub repository.

Please open a ticket in the `corporate-it` on slack with this request.

## Developing WKP

### Dependencies

- [`hugo`](https://gohugo.io/getting-started/quick-start/) - extended version
- [`embedmd`](https://github.com/campoy/embedmd)
- [`yarn`](https://classic.yarnpkg.com/en/docs/install)

### Build

To build the `wk` binary after making changes run:

```console
make install
```

To run the unit tests:

```console
make unit-tests
```

or to run the unit tests and build all container images:

```console
make check
```

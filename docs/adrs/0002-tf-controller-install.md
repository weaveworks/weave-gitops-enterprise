# 0002. Terraform Controller Installation Mechanism

* Status: accepted
* Date: 2022-04-27
* Authors: @richardcase
* Deciders: @JamWils @bigkevmcd

## Context

With the introduction of "Terraform Templates" into Weave GitOps Enterprise (see [initiative](https://www.notion.so/weaveworks/Terraform-templates-for-WGE-1e3214ad5ff04a339cd0a2a4511e7809)) we need a way to install the [Terraform Controller](https://github.com/weaveworks/tf-controller) into the management cluster.

Not every customer will want to use Terraform Templates so whatever mechanism we choose must be optional.

The initial thought was that we would use the `gitops` cli to enable a customer to optionally do the install:

```bash
gitops install components terraform
```

However, on further discussion this was deemed not to be optimal for the following reasons:

* The install experience on the management cluster would be different to that on a tenant cluster.
* The `gitops install` command had been removed from the cli.

Giving a consistent experience of installing the Terraform Controller that is the same on the management cluster as it is on a tenant cluster was deemed to be of high importance. Some [discussion in Notion](https://www.notion.so/Terraform-templates-for-WGE-1e3214ad5ff04a339cd0a2a4511e7809?d=8e4b89f0f8f045c891a17550ae0d6395#795a3b6af5f94726adb4d95c85f8c2d7).

There are discussions still ongoing on the role of the `gitops` cli and what functionality it contains in the future.

## Decision

The installation of the Terraform Controller into the management cluster (and tenenat clusters) will be done via a `HelmRelease` that uses the projects Helm Chart (**tf-controller**) available via the [https://weaveworks.github.io/tf-controller](https://weaveworks.github.io/tf-controller) registry.

In the future we may consider using the [profile](https://github.com/weaveworks/profiles-catalog/tree/main/charts/tf-controller).

## Consequences

As a result of the decision some things to consider:

* Updating the Weave GitOps documentation with instructions on how to install the Terraform Controller via `HelmRelease`.
* Adding integration tests to Wego Enterprise to cover the TF controller installation.
* To enable functionality in the Wego UI we can check for the existence of the TF Controller CRDs and an entitlement, if both are true then show the functionality.

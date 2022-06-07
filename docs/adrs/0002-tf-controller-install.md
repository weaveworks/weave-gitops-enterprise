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

## Decision

The installation of the Terraform Controller will be done using the **Profile**.

As profiles are based on Helm charts (see [profiles ADR-0002](https://github.com/weaveworks/profiles/blob/main/docs/adrs/0002-helm-charts-as-profiles.md)) this means we can use the same source as a normal Helm chart (via a `HelmRelease`) or as a profile.

We have a profile for the [Terraform Controller](https://github.com/weaveworks/profiles-catalog/tree/main/charts/tf-controller) available. This can be used to install on the management cluster via committing a `HelmRelease` to the management clusters repo.

This profile can also be used to install the TF controller on a tenant cluster if needed via [the UI](https://docs.gitops.weave.works/docs/cluster-management/profiles/) or [the cli](https://docs.gitops.weave.works/docs/references/cli-reference/gitops_add_profile/):

```bash
gitops add profile --name=tfcontroller --cluster=tenant1 --version=1.0.0 --config-repo=ssh://git@github.com/owner/config-repo.git
```

## Consequences

We would need to do the following:

* Ensure the Weave GitOps documentation is updated with instructions on how to install the Terraform Controller
* Ensure the [Terraform Controller profile](https://github.com/weaveworks/profiles-catalog/tree/main/charts/tf-controller) is kept up to date with new releases of the controller
* Consider adding integration tests to Wego Enterprise to cover this functionality 
* To enable functionality in the Wego UI we can check for the existence of the TF Controller CRDs and an entitlement, if both are true then show the functionality.

# Weave GitOps Enterprise CI/CD

This document aims to characterise the Ci/CD pipeline for change until it reaches production.
In our context, given that we are not SaaS company, our definition of production environment
is until the change gets released as part of a [Weave GitOps release](https://github.com/weaveworks/weave-gitops-enterprise/releases)
As expected, Flux for deployment and Github Actions as CI.

## PR Journey 

The journey of a Weaveworks engineer development is the following:

1. A developer works in a feature in a feature branch until ready to review 
2. Raises a Pull Request that triggers the [PR github action workflow](../.github/workflows/test.yaml) that:
   - lints and code standards
   - run unit test for frontend and backend
   - run integration tests for backend
   - build and publish artifacts: container images, npm packages, binaries and helm charts
3. Pull Request is reviewed
4. When approved and merged, `deploy` [github workflow](../.github/workflows/deploy) is executed that:
   - run build and test steps for the integrated code
   - run a set of e2e or smoke tests to ensure the app is health
5. Once passes CI for main and artifacts are built and pushed. The feature is then 
deployed to [Staging](https://gitops.internal-dev.wego-gke.weave.works) environment by Flux. This environment helps us to:
    - Functional testing of the feature in a released environment
    - Monitor performance of the capability
6. The feature will be released according to the week-release cadence that we have for Weave GitOps.

## Environments

In the context of environments like dev, test, prod, our picture is the following:

- Dev: happens within a developer's machine where a range of unit and integration test supports the process, as well as, [tilt](https://tilt.dev/) 
that we use to recreate the application locally. See [Tiltfile](../Tiltfile).
- Test/Staging: once code gets into main branch, it gets deployed to [Staging](https://gitops.internal-dev.wego-gke.weave.works) environment where a developer 
is able to monitor its behaviour in a long-lived environment. 
- Production: we don't have a production environment as compared to a SaaS company, as our product gets deployed by customers in their environments. However, we have internal 
customers (ex. Sales) that provide us early-feedback for any new Weave GitOps Enterprise release.  

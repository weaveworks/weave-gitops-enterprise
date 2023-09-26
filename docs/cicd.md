# Weave Gitops Enterprise CI/CD

Weave Gitops CICD systems is based on, as expected, Flux for deployment and Github Actions as CI. 

The journey of a weaveworks engineer development is the following:

1. A developer works in a feature in a feature branch until ready to review 
2. Raises a Pull Request that triggers the [PR github action workflow](../.github/workflows/test.yaml) that:
   - lints and code standards
   - run unit test for frontend and backend
   - run integration tests for backend
   - build and publish artifacts: container images, npm packages, binaries and helm charts
3. Pull Request is reviewed
4. When approved and merged, another github workflow is executed that:
   - run build and test steps for the integrated code
   - run a set of e2e or smoker tests to ensure the app is health
5. Once passes CI for main and artifacts are built and pushed. The release candidate is then 
deployed to [Staging](https://gitops.internal-dev.wego-gke.weave.works) environment by Flux. This environment helps us to:
    - Functional testing of the feature in a released environment
    - Monitor performance of the capability
6. The feature will be released according to the week-release cadence that we have for Weave Gitops.


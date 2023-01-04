# Workflows

##  Check that PR has a label for use in release notes

The WGE release notes are generated from PR titles where each title gets added to one of those four section: Enhancements, UI, Bugs, and Tests. The `Check that PR has a label for use in release notes` workflow, which runs on open PRs, is used to check that the PR has a label to indicate which section it should be added to when the release notes are generated.

Pull requests require exactly one label from the allowed labels:

 1. `ui`: New feature or request in the UI
 1. `enhancement`: New feature or request in the BE
 2. `bug`: Bug fixes
 3. `test`: Mark a PR as being about tests
 4. `exclude from release notes`: Use this label to exclude a PR from the release notes ex: doc changes

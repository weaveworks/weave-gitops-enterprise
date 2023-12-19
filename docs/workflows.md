# Workflows

##  Check that PR has a label for use in release notes

The WGE release notes are generated from PR titles where each title gets added to one of those four section: Enhancements, UI, Bugs, and Tests. The `Check that PR has a label for use in release notes` workflow, which runs on open PRs, is used to check that the PR has a label to indicate which section it should be added to when the release notes are generated.

Pull requests require exactly one label from the allowed labels:

 1. `ui`: New feature or request in the UI
 2. `enhancement`: New feature or request in the BE
 3. `dependencies`: Use this label for dependency management PRs (i.e. dependabot). 
 4. `bug`: Bug fixes
 5. `test`: Mark a PR as being about tests
 6. `exclude from release notes`: Use this label to exclude a PR from the release notes ex: doc changes

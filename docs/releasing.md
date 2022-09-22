# Releasing `weave-gitops-enterprise`

[comment]: <> (Github can generate TOCs now see https://github.blog/changelog/2021-04-13-table-of-contents-support-in-markdown-files/)

How to release a new version of weave-gitops-enterprise

## Prerequisites

Install [GnuPG](https://gnupg.org/) and [generate a GPG key and add it to your Github account](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key).

If you aren't using the GPG suite, you will need to [add the GPG key to your .bashrc / .zshrc](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key).

## Checklist

### Radiate that a release is happening (**30 mins** waiting for objections)

Write a message in `#weave-gitops-dev` on slack. Ideally we want to fall into a predicatable release cadence so that the release doesn't come as a surprise to anyone.

Wait and see if any team has an objection like a known release blocking bug.

### Make sure weave-gitops-enterprise on `main` is green (**30s** to check)

Look for the green tick next to the last commit on [weave-gitops-enterprise](https://github.com/weaveworks/weave-gitops-enterprise)

### Make sure dependencies are up to date (**5 mins** to check. **30 mins** to correct versions and wait for green `main`)

In particular:
- weave-gitops ([releases](https://github.com/weaveworks/weave-gitops/releases))
  - https://github.com/weaveworks/weave-gitops-enterprise/blob/main/go.mod
  - https://github.com/weaveworks/weave-gitops-enterprise/blob/main/ui-cra/package.json
- cluster-controller ([releases](https://github.com/weaveworks/cluster-controller/releases))
  - https://github.com/weaveworks/weave-gitops-enterprise/blob/main/go.mod
  - https://github.com/weaveworks/weave-gitops-enterprise/blob/main/charts/cluster-controller/values.yaml
- cluster-bootstrap-controller ([releases](https://github.com/weaveworks/cluster-bootstrap-controller/releases))
  - https://github.com/weaveworks/weave-gitops-enterprise/blob/main/charts/mccp/values.yaml

Check [how to update things in ../CONTRIBUTING.md](../CONTRIBUTING.md#how-to-update-the-version-of-weave-gitops) for instructions on how to update properly.

## Release
### Create a tag

Tag the new release with an **annotated** and **signed** tag.

```bash
cd weave-gitops-enterprise
git checkout main
git pull

# Make sure your local commit is the same as the head on github
git log 

# Tag a push an annotated tag
git tag -a -s v0.9.4 -m "Weave GitOps Enterprise v0.9.4"
git push origin v0.9.4
```

This will kick off the release process in GitHub actions
- View progress via the [release action](https://github.com/weaveworks/weave-gitops-enterprise/actions/workflows/release.yaml).
- The action will generate release notes and publish a release on GitHub
- A message will be sent to #weave-gitops-dev in slack announcing the release

### Update the release notes with *Dependencies* and *Highlights*

You can use a [previous release](https://github.com/weaveworks/weave-gitops-enterprise/releases) as a template here.

Edit the new release, copy and paste the _Dependency Versions_ section from an older releases. Update any versions that have been changed.

Add a Highlights section, calling out the significant changes in this new release.

